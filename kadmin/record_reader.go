package kadmin

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"ktea/serdes"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/log"

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

type StartPoint int64

const (
	Beginning StartPoint = iota
	MostRecent
	Today
	Yesterday
	Last7Days
	Live
)

func (p *StartPoint) time() int64 {
	switch *p {
	case Beginning, MostRecent, Live:
		return sarama.OffsetOldest
	case Today:
		t := time.Now()
		startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		return startOfDay.UnixMilli()
	case Yesterday:
		t := time.Now().AddDate(0, 0, -1)
		startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		return startOfDay.UnixMilli()
	case Last7Days:
		t := time.Now().AddDate(0, 0, -7)
		startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		return startOfDay.UnixMilli()
	}

	return int64(*p)
}

type RecordReader interface {
	ReadRecords(ctx context.Context, rd ReadDetails) tea.Msg
}

type ReadingStartedMsg struct {
	ConsumerRecord chan ConsumerRecord
	EmptyTopic     chan bool
	// NoRecordsFound indicates that there are no records in any of the selected partitions
	// for the given filter criteria.
	NoRecordsFound chan bool
	Err            chan error
	CancelFunc     context.CancelFunc
}

func (m *ReadingStartedMsg) AwaitRecord() tea.Msg {
	select {
	case record, ok := <-m.ConsumerRecord:
		if !ok {
			return ConsumptionEndedMsg{}
		}

		return ConsumerRecordReceived{
			Records:        []ConsumerRecord{record},
			consumerRecord: m.ConsumerRecord,
			emptyTopic:     m.EmptyTopic,
			noRecordsFound: m.NoRecordsFound,
			err:            m.Err,
			cancelFunc:     m.CancelFunc,
		}
	case empty := <-m.EmptyTopic:
		if empty {
			return EmptyTopicMsg{}
		}
		return nil
	case noRecords := <-m.NoRecordsFound:
		if noRecords {
			return NoRecordsFound{
				consumerRecord: m.ConsumerRecord,
				emptyTopic:     m.EmptyTopic,
				noRecordsFound: m.NoRecordsFound,
				err:            m.Err,
				cancelFunc:     m.CancelFunc,
			}
		}
		return nil
	case err := <-m.Err:
		return err
	}
}

func (m *ReadingStartedMsg) shutdown() {
	m.CancelFunc()
	close(m.ConsumerRecord)
	close(m.Err)
	close(m.EmptyTopic)
	close(m.NoRecordsFound)
}

type ConsumerRecordReceived struct {
	Records        []ConsumerRecord
	consumerRecord chan ConsumerRecord
	emptyTopic     chan bool
	noRecordsFound chan bool
	err            chan error
	cancelFunc     context.CancelFunc
}

func (m *ConsumerRecordReceived) AwaitNextRecord() tea.Msg {
	log.Debug("awaiting next record")
	select {
	case record, ok := <-m.consumerRecord:
		if !ok {
			return ConsumptionEndedMsg{}
		}

		records := []ConsumerRecord{record}
		timeout := time.After(50 * time.Millisecond)
		for {
			select {
			case r := <-m.consumerRecord:
				records = append(records, r)
			case <-timeout:
				return ConsumerRecordReceived{
					Records:        records,
					consumerRecord: m.consumerRecord,
					emptyTopic:     m.emptyTopic,
					noRecordsFound: m.noRecordsFound,
					err:            m.err,
					cancelFunc:     m.cancelFunc,
				}
			}
		}

	case emptyTopic := <-m.emptyTopic:
		if emptyTopic {
			return EmptyTopicMsg{}
		}
		return nil
	case err := <-m.err:
		return err
	}
}

type EmptyTopicMsg struct {
}

type NoRecordsFound struct {
	consumerRecord chan ConsumerRecord
	emptyTopic     chan bool
	noRecordsFound chan bool
	err            chan error
	cancelFunc     context.CancelFunc
}

type ConsumptionEndedMsg struct{}

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

func NewDefaultReadDetails(topic *ListedTopic) ReadDetails {
	return ReadDetails{
		TopicName:       topic.Name,
		PartitionToRead: topic.Partitions(),
		StartPoint:      MostRecent,
		Limit:           500,
		Filter:          &Filter{},
	}
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

func (record *ConsumerRecord) PayloadType() string {
	// schema is not empty, so it's Avro
	if record.Payload.Schema != "" {
		return "Avro"
	}

	// value is empty, so it's plain text'
	value := strings.TrimSpace(record.Payload.Value)
	if value == "" {
		return "Plain Text"
	}

	// value is a valid json, so it's a json'
	if json.Valid([]byte(value)) {
		return "Plain Json"
	}

	// value is a valid xml, so it's a xml'
	if isValidXML([]byte(value)) {
		return "Plain XML"
	}

	// default value is plain text
	return "Plain Text"
}

func isValidXML(data []byte) bool {
	err := xml.Unmarshal(data, new(interface{}))
	return err == nil
}

type offsets struct {
	start int64
	// most recent available, unused, offset
	end int64
}

func (o *offsets) newest() int64 {
	return o.end - 1
}

func (ka *SaramaKafkaAdmin) ReadRecords(ctx context.Context, rd ReadDetails) tea.Msg {
	ctx, cancelFunc := context.WithCancel(ctx)
	startedMsg := &ReadingStartedMsg{
		ConsumerRecord: make(chan ConsumerRecord, len(rd.PartitionToRead)),
		Err:            make(chan error, 1),
		EmptyTopic:     make(chan bool, 1),
		NoRecordsFound: make(chan bool, 1),
		CancelFunc:     cancelFunc,
	}

	go ka.doReadRecords(ctx, rd, startedMsg, cancelFunc)
	return startedMsg
}

func (ka *SaramaKafkaAdmin) doReadRecords(
	ctx context.Context,
	rd ReadDetails,
	startedMsg *ReadingStartedMsg,
	cancelFunc context.CancelFunc,
) {
	client, err := sarama.NewConsumerFromClient(ka.client)
	if err != nil {
		startedMsg.shutdown()
	}

	var (
		msgCount atomic.Int64
		wg       sync.WaitGroup
		offsets  map[int]offsets
	)

	offsets, err = ka.fetchOffsets(rd.PartitionToRead, rd.TopicName, rd.StartPoint)
	if err != nil {
		startedMsg.Err <- err
		close(startedMsg.ConsumerRecord)
		close(startedMsg.Err)
		cancelFunc()
	}

	if noRecordsFound(offsets) {
		cancelFunc()
		startedMsg.NoRecordsFound <- true
		return
	}

	emptyTopic := true

	log.Debug("Starting to read records",
		"partition", rd.PartitionToRead,
		"offsets", offsets)

	for _, p := range rd.PartitionToRead {
		// if there is no data in the partition, we don't need to read it unless live consumption is requested
		partition := p
		if offsets[partition].end != offsets[partition].start || rd.StartPoint == Live {

			emptyTopic = false

			wg.Go(func() {
				readingOffsets := ka.determineReadingOffsets(rd, offsets[partition])
				log.Debug("Reading offsets determined",
					"topic", rd.TopicName,
					"partition", partition,
					"start", readingOffsets.start,
					"end", readingOffsets.end,
				)

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

						if msgCount.Add(1) >= int64(rd.Limit) {
							select {
							case startedMsg.ConsumerRecord <- consumerRecord:
							case <-ctx.Done():
							}
							// Now that the last message is sent (or we're exiting), return.
							return
						}

						select {
						case startedMsg.ConsumerRecord <- consumerRecord:
						case <-ctx.Done():
							return
						}

						if msg.Offset == readingOffsets.end && rd.StartPoint != Live {
							return
						}
					}
				}
			})
		}
	}

	if emptyTopic {
		startedMsg.EmptyTopic <- true
	}

	go func() {
		wg.Wait()
		time.Sleep(50 * time.Millisecond)
		startedMsg.shutdown()
	}()
}

func noRecordsFound(offsets map[int]offsets) bool {
	for _, off := range offsets {
		// -1 indicates that no records exist for the requested offsets
		if off.start != -1 {
			return false
		}
	}
	return true
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
			start: offsets.end,
			end:   -1,
		}
	}

	var startOffset int64
	var endOffset int64
	numberOfRecordsPerPart := int64(float64(int64(rd.Limit)) / float64(len(rd.PartitionToRead)))
	if rd.StartPoint == Beginning {
		startOffset, endOffset = ka.determineOffsetsFromBeginning(
			offsets,
			numberOfRecordsPerPart,
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
	if startOffset < 0 || startOffset < offsets.start {
		startOffset = offsets.start
	}
	return startOffset, endOffset
}

func (ka *SaramaKafkaAdmin) determineOffsetsFromBeginning(
	offsets offsets,
	numberOfRecordsPerPart int64,
) (int64, int64) {
	var (
		startOffset int64
		endOffset   int64
	)
	startOffset = offsets.start
	if (offsets.start + numberOfRecordsPerPart) < offsets.newest() {
		endOffset = startOffset + numberOfRecordsPerPart - 1
	} else {
		endOffset = offsets.newest()
	}
	return startOffset, endOffset
}

func (ka *SaramaKafkaAdmin) fetchOffsets(
	partitions []int,
	topicName string,
	startPoint StartPoint,
) (map[int]offsets, error) {
	offsetsByPartition := make(map[int]offsets)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorsChan := make(chan error, len(partitions))

	for _, p := range partitions {

		partition := p

		log.Debug("Fetching offsets", "topic", topicName, "partition", partition)

		wg.Go(func() {
			startOffset, err := ka.client.GetOffset(
				topicName,
				int32(partition),
				startPoint.time(),
			)
			if err != nil {
				errorsChan <- err
				return
			}
			log.Debug(
				"Fetched start offset",
				"topic", topicName,
				"partition", partition,
				"startOffset", startOffset,
			)

			endOffset, err := ka.client.GetOffset(
				topicName,
				int32(partition),
				sarama.OffsetNewest,
			)
			if err != nil {
				errorsChan <- err
				return
			}
			log.Debug("Fetched end offset",
				"topic", topicName,
				"partition", partition,
				"endOffset", endOffset)

			mu.Lock()
			offsetsByPartition[partition] = offsets{
				startOffset,
				endOffset,
			}
			mu.Unlock()
		})
	}

	wg.Wait()

	select {
	case err := <-errorsChan:
		return nil, err
	default:
		return offsetsByPartition, nil
	}
}
