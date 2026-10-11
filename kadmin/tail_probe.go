package kadmin

import (
	"errors"

	"github.com/IBM/sarama"
)

const (
	probeInitialWindow int64 = 64
	probeMaxBytes      int32 = 1 << 20
)

var errProbeUnsupported = errors.New("tail probe not supported for legacy message format")

// lastDataOffset returns the highest offset in [start, end) that holds a record a consumer
// actually receives, or -1 when there is none. The newest offset of a partition is not
// necessarily such a record, e.g. transaction control markers are never delivered.
func (ka *SaramaKafkaAdmin) lastDataOffset(
	topic string,
	partition int32,
	start int64,
	end int64,
) (int64, error) {
	hi := end
	window := probeInitialWindow
	for hi > start {
		from := max(start, hi-window)
		last, err := ka.scanLastDataOffset(topic, partition, from, hi)
		if err != nil {
			return -1, err
		}
		if last >= 0 {
			return last, nil
		}
		hi = from
		window *= 4
	}
	return -1, nil
}

// scanLastDataOffset fetches [from, hi) and returns the highest offset of a delivered record in it, or -1.
func (ka *SaramaKafkaAdmin) scanLastDataOffset(
	topic string,
	partition int32,
	from int64,
	hi int64,
) (int64, error) {
	broker, err := ka.client.Leader(topic, partition)
	if err != nil {
		return -1, err
	}

	conf := ka.client.Config()
	last := int64(-1)
	offset := from
	for offset < hi {
		request, err := newProbeFetchRequest(conf)
		if err != nil {
			return -1, err
		}
		request.AddBlock(topic, partition, offset, probeMaxBytes, -1)

		response, err := broker.Fetch(request)
		if err != nil {
			return -1, err
		}
		block := response.GetBlock(topic, partition)
		if block == nil {
			return -1, sarama.ErrIncompleteResponse
		}
		if !errors.Is(block.Err, sarama.ErrNoError) {
			return -1, block.Err
		}

		next := offset
		aborted := make(map[int64]struct{})
		abortedTxns := block.AbortedTransactions
		for _, records := range block.RecordsSet {
			batch := records.RecordBatch
			if batch == nil {
				return -1, errProbeUnsupported
			}
			if batch.PartialTrailingRecord {
				break
			}
			next = max(next, batch.LastOffset()+1)

			for len(abortedTxns) > 0 && abortedTxns[0].FirstOffset <= batch.LastOffset() {
				aborted[abortedTxns[0].ProducerID] = struct{}{}
				abortedTxns = abortedTxns[1:]
			}

			if batch.Control {
				delete(aborted, batch.ProducerID)
				continue
			}
			if conf.Consumer.IsolationLevel == sarama.ReadCommitted && batch.IsTransactional {
				if _, isAborted := aborted[batch.ProducerID]; isAborted {
					continue
				}
			}

			for _, rec := range batch.Records {
				o := batch.FirstOffset + rec.OffsetDelta
				if o >= from && o < hi && o > last {
					last = o
				}
			}
		}

		if next == offset {
			// nothing (more) returned
			break
		}
		offset = next
	}

	return last, nil
}

func newProbeFetchRequest(conf *sarama.Config) (*sarama.FetchRequest, error) {
	if !conf.Version.IsAtLeast(sarama.V0_11_0_0) {
		return nil, errProbeUnsupported
	}
	request := &sarama.FetchRequest{
		Version:     5,
		MinBytes:    1,
		MaxWaitTime: 0,
		MaxBytes:    probeMaxBytes,
		Isolation:   conf.Consumer.IsolationLevel,
	}
	if conf.Version.IsAtLeast(sarama.V1_1_0_0) {
		request.Version = 7
		request.SessionID = 0
		request.SessionEpoch = -1
	}
	return request, nil
}
