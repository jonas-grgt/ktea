package kadmin

import (
	"context"
	"github.com/IBM/sarama"
	tea "github.com/charmbracelet/bubbletea"
	"ktea/serdes"
	"sync"
	"time"
)

type StartPoint int

const (
	Beginning  StartPoint = 0
	MostRecent StartPoint = 1
)

type RecordReader interface {
	ReadRecords(ctx context.Context, rd ReadDetails) tea.Msg
}

type ReadingStartedMsg struct {
	ConsumerRecord chan ConsumerRecord
	Err            chan error
}

type ReadDetails struct {
	Topic      *Topic
	Partitions []int
	StartPoint StartPoint
	Limit      int
}

type Header struct {
	Key   string
	Value string
}

type ConsumerRecord struct {
	Key       string
	Value     string
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

type EmptyTopicMsg struct {
}

func (ka *SaramaKafkaAdmin) ReadRecords(ctx context.Context, rd ReadDetails) tea.Msg {
	ctx, cancelFunc := context.WithCancel(ctx)
	startedMsg := ReadingStartedMsg{
		ConsumerRecord: make(chan ConsumerRecord, len(rd.Partitions)),
		Err:            make(chan error),
	}

	client, err := sarama.NewConsumerFromClient(ka.client)
	if err != nil {
		close(startedMsg.ConsumerRecord)
		close(startedMsg.Err)
		return startedMsg
	}

	var (
		msgCount   int
		mu         sync.Mutex
		closeOnce  sync.Once
		wg         sync.WaitGroup
		offsets    map[int]offsets
		ok         bool
		partitions []int
	)

	partitions = ka.determineReadPartitions(rd)

	offsets, ok = ka.fetchOffsets(partitions, rd, startedMsg)
	if !ok {
		close(startedMsg.Err)
		cancelFunc()
		return startedMsg
	}

	wg.Add(len(partitions))

	var atLeastOnePartitionReadable bool
	for _, partition := range partitions {
		if offsets[partition].firstAvailable != offsets[partition].oldest {
			atLeastOnePartitionReadable = true
			go func(partition int) {
				defer wg.Done()

				readingOffsets := ka.determineReadingOffsets(rd, offsets[partition])
				consumer, err := client.ConsumePartition(
					rd.Topic.Name,
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
								string(h.Value),
							})
						}

						consumerRecord := ConsumerRecord{
							Key:       string(msg.Key),
							Value:     ka.deserialize(err, msg),
							Partition: int64(msg.Partition),
							Offset:    msg.Offset,
							Headers:   headers,
							Timestamp: msg.Timestamp,
						}

						var shouldClose bool

						mu.Lock()
						msgCount++
						if msgCount >= rd.Limit {
							shouldClose = true
						}
						mu.Unlock()

						select {
						case startedMsg.ConsumerRecord <- consumerRecord:
						case <-ctx.Done():
							return
						}

						if shouldClose {
							cancelFunc() // Cancel the context to stop other goroutines
							return
						}

						if msg.Offset == readingOffsets.end {
							return
						}
					}
				}
			}(partition)
		}
	}

	go func() {
		wg.Wait()
		closeOnce.Do(func() {
			close(startedMsg.ConsumerRecord)
		})
	}()

	if atLeastOnePartitionReadable {
		return startedMsg
	} else {
		return EmptyTopicMsg{}
	}
}

func (ka *SaramaKafkaAdmin) deserialize(
	err error,
	msg *sarama.ConsumerMessage,
) string {
	deserializer := serdes.NewAvroDeserializer(ka.sra)
	var payload string
	payload, err = deserializer.Deserialize(msg.Value)
	if err != nil {
		payload = err.Error()
	}
	return payload
}

func (ka *SaramaKafkaAdmin) determineReadPartitions(rd ReadDetails) []int {
	var partitions []int
	if len(rd.Partitions) == 0 {
		partitions = make([]int, rd.Topic.Partitions)
		for i := range partitions {
			partitions[i] = i
		}
	} else {
		partitions = rd.Partitions
	}
	return partitions
}

type readingOffsets struct {
	start int64
	end   int64
}

func (ka *SaramaKafkaAdmin) determineReadingOffsets(
	rd ReadDetails,
	offsets offsets,
) readingOffsets {
	var startOffset int64
	var endOffset int64
	numberOfRecordsPerPart := int64(float64(int64(rd.Limit)) / float64(rd.Topic.Partitions))
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
	rd ReadDetails,
	rsm ReadingStartedMsg,
) (map[int]offsets, bool) {
	offsetsByPartition := make(map[int]offsets)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorsChan := make(chan error, len(partitions))

	for _, partition := range partitions {
		wg.Add(1)
		go func(partition int) {
			defer wg.Done()

			firstAvailableOffset, err := ka.client.GetOffset(
				rd.Topic.Name,
				int32(partition),
				sarama.OffsetNewest,
			)
			if err != nil {
				errorsChan <- err
				return
			}

			oldestOffset, err := ka.client.GetOffset(
				rd.Topic.Name,
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
		rsm.Err <- err
		close(rsm.ConsumerRecord)
		close(rsm.Err)
		return nil, false
	default:
		return offsetsByPartition, true
	}
}
