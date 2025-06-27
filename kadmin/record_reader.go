package kadmin

import (
	"bytes"
	"context"
	"encoding/binary"
	"github.com/charmbracelet/log"
	"ktea/serdes"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/IBM/sarama"
	tea "github.com/charmbracelet/bubbletea"
)

type FilterType string

func (filterDetails *Filter) Filter(value string) bool {
	switch filterDetails.KeyFilter {
	case ContainsFilterType:
		return strings.Contains(value, filterDetails.KeySearchTerm)
	case StartsWithFilterType:
		return strings.HasPrefix(value, filterDetails.KeySearchTerm)
	default:
		return true
	}
}

const (
	ContainsFilterType   FilterType = "contains"
	StartsWithFilterType FilterType = "starts with"
	NoFilterType         FilterType = "none"
)

type StartPoint int

const (
	Beginning  StartPoint = 0
	MostRecent StartPoint = 1
	Live       StartPoint = 2
)

type RecordReader interface {
	ReadRecords(ctx context.Context, rd ReadDetails) tea.Msg
}

type ReadingStartedMsg struct {
	ConsumerRecord chan ConsumerRecord
	EmptyTopic     chan bool
	Err            chan error
	CancelFunc     context.CancelFunc
}

type Filter struct {
	KeyFilter       FilterType
	KeySearchTerm   string
	ValueFilter     FilterType
	ValueSearchTerm string
}

type ReadDetails struct {
	TopicName       string
	PartitionToRead []int
	StartPoint      StartPoint
	Limit           int
	Filter          *Filter
}

type HeaderValue struct {
	data []byte
}

func NewHeaderValue(data string) HeaderValue {
	return HeaderValue{[]byte(data)}
}

func (v HeaderValue) String() string {
	if utf8.Valid(v.data) {
		return string(v.data)
	}

	if len(v.data) >= 4 {
		var int32Val int32
		err := binary.Read(bytes.NewReader(v.data), binary.BigEndian, &int32Val)
		if err == nil {
			return string(int32Val)
		}
	}
	if len(v.data) >= 8 {
		var int64Val int64
		err := binary.Read(bytes.NewReader(v.data), binary.BigEndian, &int64Val)
		if err == nil {
			return strconv.FormatInt(int64Val, 10)
		}
	}

	if len(v.data) >= 4 {
		var float32Val float32
		err := binary.Read(bytes.NewReader(v.data), binary.BigEndian, &float32Val)
		if err == nil {
			return strconv.FormatFloat(float64(float32Val), 'f', -1, 32)
		}
	}
	if len(v.data) >= 8 {
		var float64Val float64
		err := binary.Read(bytes.NewReader(v.data), binary.BigEndian, &float64Val)
		if err == nil {
			return strconv.FormatFloat(float64Val, 'f', -1, 64)
		}
	}

	return string(v.data)
}

type Header struct {
	Key   string
	Value HeaderValue
}

type ConsumerRecord struct {
	Key       string
	Payload   serdes.DesData
	Err       error
	Partition int64
	Offset    int64
	Headers   []Header
	Timestamp time.Time
}

type offsets struct {
	oldest int64
	// most recent available, unused, offset
	firstAvailable int64
}

func (o *offsets) newest() int64 {
	return o.firstAvailable - 1
}

func (ka *SaramaKafkaAdmin) ReadRecords(ctx context.Context, rd ReadDetails) tea.Msg {
	ctx, cancelFunc := context.WithCancel(ctx)
	startedMsg := ReadingStartedMsg{
		ConsumerRecord: make(chan ConsumerRecord, len(rd.PartitionToRead)),
		Err:            make(chan error),
		EmptyTopic:     make(chan bool),
		CancelFunc:     cancelFunc,
	}

	go ka.doReadRecords(ctx, rd, startedMsg, cancelFunc)
	return startedMsg
}

func (ka *SaramaKafkaAdmin) doReadRecords(
	ctx context.Context,
	rd ReadDetails,
	startedMsg ReadingStartedMsg,
	cancelFunc context.CancelFunc,
) {
	client, err := sarama.NewConsumerFromClient(ka.client)
	if err != nil {
		close(startedMsg.ConsumerRecord)
		close(startedMsg.Err)
	}

	var (
		msgCount  atomic.Int64
		closeOnce sync.Once
		wg        sync.WaitGroup
		offsets   map[int]offsets
	)

	offsets, err = ka.fetchOffsets(rd.PartitionToRead, rd.TopicName)
	if err != nil {
		startedMsg.Err <- err
		close(startedMsg.ConsumerRecord)
		close(startedMsg.Err)
		cancelFunc()
	}

	wg.Add(len(rd.PartitionToRead))

	emptyTopic := true
	for _, partition := range rd.PartitionToRead {
		// if there is no data in the partition, we don't need to read it unless live consumption is requested
		if offsets[partition].firstAvailable != offsets[partition].oldest || rd.StartPoint == Live {
			emptyTopic = false
			go func(partition int) {
				defer wg.Done()

				readingOffsets := ka.determineReadingOffsets(rd, offsets[partition])
				consumer, err := client.ConsumePartition(
					rd.TopicName,
					int32(partition),
					readingOffsets.start,
				)
				if err != nil {
					startedMsg.Err <- err
					cancelFunc()
					return
				}
				defer consumer.Close()

				msgChan := consumer.Messages()

				for {
					select {
					case err := <-consumer.Errors():
						startedMsg.Err <- err
						return
					case <-ctx.Done():
						return
					case msg := <-msgChan:
						var headers []Header
						for _, h := range msg.Headers {
							headers = append(headers, Header{
								string(h.Key),
								HeaderValue{h.Value},
							})
						}

						var desData serdes.DesData
						key := string(msg.Key)
						desData, err = ka.deserialize(msg)

						if rd.Filter != nil && err == nil {
							if !ka.matchesFilter(key, desData.Value, rd.Filter) {
								continue
							}
						}

						consumerRecord := ConsumerRecord{
							Key:       key,
							Payload:   desData,
							Err:       err,
							Partition: int64(msg.Partition),
							Offset:    msg.Offset,
							Headers:   headers,
							Timestamp: msg.Timestamp,
						}

						var shouldClose bool

						if msgCount.Add(1) >= int64(rd.Limit) {
							shouldClose = true
						}

						select {
						case startedMsg.ConsumerRecord <- consumerRecord:
						case <-ctx.Done():
							return
						}

						if shouldClose {
							cancelFunc() // Cancel the context to stop other goroutines
							return
						}

						if msg.Offset == readingOffsets.end && rd.StartPoint != Live {
							return
						}
					}
				}
			}(partition)
		}
	}

	if emptyTopic {
		cancelFunc()
		startedMsg.EmptyTopic <- true
	}

	go func() {
		wg.Wait()
		closeOnce.Do(func() {
			close(startedMsg.ConsumerRecord)
			close(startedMsg.Err)
		})
	}()
}

func (ka *SaramaKafkaAdmin) matchesFilter(key, value string, filterDetails *Filter) bool {
	if filterDetails == nil {
		return true
	}

	if filterDetails.KeyFilter != NoFilterType {
		return filterDetails.Filter(key)
	}

	if filterDetails.ValueSearchTerm != "" && !strings.Contains(value, filterDetails.ValueSearchTerm) {
		return false
	}

	return true
}

func (ka *SaramaKafkaAdmin) deserialize(
	msg *sarama.ConsumerMessage,
) (serdes.DesData, error) {
	deserializer := serdes.NewAvroDeserializer(ka.sra)
	return deserializer.Deserialize(msg.Value)
}

type readingOffsets struct {
	start int64
	end   int64
}

func (ka *SaramaKafkaAdmin) determineReadingOffsets(
	rd ReadDetails,
	offsets offsets,
) readingOffsets {

	if rd.StartPoint == Live {
		return readingOffsets{
			start: offsets.firstAvailable,
			end:   -1,
		}
	}

	var startOffset int64
	var endOffset int64
	numberOfRecordsPerPart := int64(float64(int64(rd.Limit)) / float64(len(rd.PartitionToRead)))
	if rd.StartPoint == Beginning {
		startOffset, endOffset = ka.determineOffsetsFromBeginning(
			startOffset,
			offsets,
			numberOfRecordsPerPart,
			endOffset,
		)
	} else {
		startOffset, endOffset = ka.determineMostRecentOffsets(
			startOffset,
			offsets,
			numberOfRecordsPerPart,
			endOffset,
		)
	}
	return readingOffsets{
		start: startOffset,
		end:   endOffset,
	}
}

func (ka *SaramaKafkaAdmin) determineMostRecentOffsets(
	startOffset int64,
	offsets offsets,
	numberOfRecordsPerPart int64,
	endOffset int64,
) (int64, int64) {
	startOffset = offsets.newest() - numberOfRecordsPerPart
	endOffset = offsets.newest()
	if startOffset < 0 || startOffset < offsets.oldest {
		startOffset = offsets.oldest
	}
	return startOffset, endOffset
}

func (ka *SaramaKafkaAdmin) determineOffsetsFromBeginning(
	startOffset int64,
	offsets offsets,
	numberOfRecordsPerPart int64,
	endOffset int64,
) (int64, int64) {
	startOffset = offsets.oldest
	if offsets.oldest+numberOfRecordsPerPart < offsets.newest() {
		endOffset = startOffset + numberOfRecordsPerPart - 1
	} else {
		endOffset = offsets.newest()
	}
	return startOffset, endOffset
}

func (ka *SaramaKafkaAdmin) fetchOffsets(
	partitions []int,
	topicName string,
) (map[int]offsets, error) {
	offsetsByPartition := make(map[int]offsets)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorsChan := make(chan error, len(partitions))

	for _, partition := range partitions {

		log.Debug("fetching offsets", "topic", topicName, "partition", partition)

		wg.Add(1)
		go func(partition int) {
			defer wg.Done()

			firstAvailableOffset, err := ka.client.GetOffset(
				topicName,
				int32(partition),
				sarama.OffsetNewest,
			)
			if err != nil {
				errorsChan <- err
				return
			}

			oldestOffset, err := ka.client.GetOffset(
				topicName,
				int32(partition),
				sarama.OffsetOldest,
			)
			if err != nil {
				errorsChan <- err
				return
			}

			mu.Lock()
			offsetsByPartition[partition] = offsets{
				oldestOffset,
				firstAvailableOffset,
			}
			mu.Unlock()
		}(partition)
	}

	wg.Wait()

	select {
	case err := <-errorsChan:
		return nil, err
	default:
		return offsetsByPartition, nil
	}
}
