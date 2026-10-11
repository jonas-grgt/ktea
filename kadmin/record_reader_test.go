package kadmin

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
	"ktea/serdes"
	"strconv"
	"testing"
	"time"
)

func TestReadRecords(t *testing.T) {
	t.Run("Empty topic", func(t *testing.T) {
		topic := topicName()
		// given
		msg := ka.CreateTopic(TopicCreationDetails{
			Name:              topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}).(TopicCreationStartedMsg)

		switch msg.AwaitCompletion().(type) {
		case TopicCreatedMsg:
		case TopicCreationErrMsg:
			t.Fatal("Unable to create topic", msg.Err)
		}
		// when
		rsm := ka.ReadRecords(context.Background(), ReadDetails{
			TopicName:       topic,
			PartitionToRead: []int{0},
			StartPoint:      Beginning,
			Limit:           50,
			Filter:          nil,
		}).(*ReadingStartedMsg)

		// then
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		select {
		case <-rsm.EmptyTopic:
		case <-ctx.Done():
			t.Fatal("timed out waiting for EmptyTopic signal")
		}

		// clean up
		ka.DeleteTopic(topic)
	})

	t.Run("Read from beginning with specific limit", func(t *testing.T) {
		t.Run("with one partition", func(t *testing.T) {
			topic := topicName()
			// given
			msg := ka.CreateTopic(TopicCreationDetails{
				Name:              topic,
				NumPartitions:     1,
				ReplicationFactor: 1,
			}).(TopicCreationStartedMsg)

			switch msg.AwaitCompletion().(type) {
			case TopicCreatedMsg:
			case TopicCreationErrMsg:
				t.Fatal("Unable to create topic", msg.Err)
			}
			// when
			assert.EventuallyWithT(t, func(c *assert.CollectT) {
				for i := 0; i < 55; i++ {
					psm := ka.PublishRecord(&ProducerRecord{
						Topic: topic,
						Key:   strconv.Itoa(i),
						Value: []byte("{\"id\":\"123\"}"),
					})

					select {
					case err := <-psm.Err:
						t.Fatal(c, "Unable to publish", err)
					case p := <-psm.Published:
						assert.True(c, p)
					}
				}
			}, 10*time.Second, 10*time.Millisecond)

			// then
			rsm := ka.ReadRecords(context.Background(), ReadDetails{
				TopicName:       topic,
				PartitionToRead: []int{0},
				StartPoint:      Beginning,
				Limit:           50,
				Filter:          nil,
			}).(*ReadingStartedMsg)

			var receivedRecords []int
			for {
				select {
				case r, ok := <-rsm.ConsumerRecord:
					if !ok {
						goto assertRecords
					}
					key, _ := strconv.Atoi(r.Key)
					receivedRecords = append(receivedRecords, key)
				}
			}

		assertRecords:
			{
				assert.Len(t, receivedRecords, 50)
				//assert.Equal(t, 49, slices.Max(receivedRecords))
				assert.Equal(t, 0, slices.Min(receivedRecords))
			}

			// clean up
			ka.DeleteTopic(topic)
		})

		t.Run("with multiple partitions", func(t *testing.T) {
			topic := topicName()
			// given
			msg := ka.CreateTopic(TopicCreationDetails{
				Name:              topic,
				NumPartitions:     4,
				ReplicationFactor: 1,
			}).(TopicCreationStartedMsg)

			switch msg.AwaitCompletion().(type) {
			case TopicCreatedMsg:
			case TopicCreationErrMsg:
				t.Fatal("Unable to create topic", msg.Err)
			}

			// when
			assert.EventuallyWithT(t, func(c *assert.CollectT) {
				for i := 0; i < 52; i++ {
					partition := i % 4
					psm := ka.PublishRecord(&ProducerRecord{
						Topic:     topic,
						Key:       strconv.Itoa(i),
						Partition: &partition,
						Value:     []byte("{\"id\":\"123\"}"),
					})

					select {
					case err := <-psm.Err:
						t.Fatal("Unable to publish", err)
					case p := <-psm.Published:
						assert.True(c, p)
					}
				}
			}, 10*time.Second, 10*time.Millisecond)

			// then
			rsm := ka.ReadRecords(context.Background(), ReadDetails{
				TopicName:       topic,
				PartitionToRead: []int{0, 1, 2, 3},
				StartPoint:      Beginning,
				Limit:           40,
			}).(*ReadingStartedMsg)

			var receivedRecords []int
			for {
				select {
				case r, ok := <-rsm.ConsumerRecord:
					if !ok {
						goto assertRecords
					}
					key, _ := strconv.Atoi(r.Key)
					receivedRecords = append(receivedRecords, key)
				}
			}

		assertRecords:
			{
				assert.Len(t, receivedRecords, 40)
				assert.Equal(t, 0, slices.Min(receivedRecords))
			}

			// clean up
			ka.DeleteTopic(topic)
		})
	})

	t.Run("Read from MostRecent", func(t *testing.T) {
		t.Run("with specific in range limit", func(t *testing.T) {
			topic := topicName()
			// given
			msg := ka.CreateTopic(TopicCreationDetails{
				Name:              topic,
				NumPartitions:     1,
				ReplicationFactor: 1,
			}).(TopicCreationStartedMsg)

			switch msg.AwaitCompletion().(type) {
			case TopicCreatedMsg:
			case TopicCreationErrMsg:
				t.Fatal("Unable to create topic", msg.Err)
			}

			// when
			assert.EventuallyWithT(t, func(c *assert.CollectT) {
				for i := 0; i < 55; i++ {
					psm := ka.PublishRecord(&ProducerRecord{
						Topic: topic,
						Key:   strconv.Itoa(i),
						Value: []byte("{\"id\":\"123\"}"),
					})

					select {
					case err := <-psm.Err:
						t.Fatal(c, "Unable to publish", err)
					case p := <-psm.Published:
						assert.True(c, p)
					}
				}
			}, 10*time.Second, 10*time.Millisecond)

			// then
			rsm := ka.ReadRecords(context.Background(), ReadDetails{
				TopicName:       topic,
				PartitionToRead: []int{0},
				StartPoint:      MostRecent,
				Limit:           50,
			}).(*ReadingStartedMsg)

			var receivedRecords []int
			for {
				select {
				case r, ok := <-rsm.ConsumerRecord:
					if !ok {
						goto assertRecords
					}
					key, _ := strconv.Atoi(r.Key)
					receivedRecords = append(receivedRecords, key)
				}
			}

		assertRecords:
			{
				assert.Equal(t, 53, slices.Max(receivedRecords))
			}

			// clean up
			ka.DeleteTopic(topic)
		})

		t.Run("with specific in out of range limit", func(t *testing.T) {
			topic := topicName()
			// given
			msg := ka.CreateTopic(TopicCreationDetails{
				Name:              topic,
				NumPartitions:     1,
				ReplicationFactor: 1,
			}).(TopicCreationStartedMsg)

			switch msg.AwaitCompletion().(type) {
			case TopicCreatedMsg:
			case TopicCreationErrMsg:
				t.Fatal("Unable to create topic", msg.Err)
			}

			// when
			assert.EventuallyWithT(t, func(c *assert.CollectT) {
				for i := 0; i < 55; i++ {
					psm := ka.PublishRecord(&ProducerRecord{
						Topic: topic,
						Key:   strconv.Itoa(i),
						Value: []byte("{\"id\":\"123\"}"),
					})

					select {
					case err := <-psm.Err:
						t.Fatal(c, "Unable to publish", err)
					case p := <-psm.Published:
						assert.True(c, p)
					}
				}
			}, 10*time.Second, 10*time.Millisecond)

			// then
			rsm := ka.ReadRecords(context.Background(), ReadDetails{
				TopicName:       topic,
				PartitionToRead: []int{0},
				StartPoint:      MostRecent,
				Limit:           500,
			}).(*ReadingStartedMsg)

			var receivedRecords []int
			for {
				select {
				case r, ok := <-rsm.ConsumerRecord:
					if !ok {
						goto assertRecords
					}
					key, _ := strconv.Atoi(r.Key)
					receivedRecords = append(receivedRecords, key)
				}
			}

		assertRecords:
			{
				assert.Equal(t, 54, slices.Max(receivedRecords))
			}

			// clean up
			ka.DeleteTopic(topic)
		})
	})

	t.Run("Read today's records", func(t *testing.T) {
		t.Run("when there are records from previous days", func(t *testing.T) {
			topic := topicName()
			// given
			msg := ka.CreateTopic(TopicCreationDetails{
				Name:              topic,
				NumPartitions:     1,
				ReplicationFactor: 1,
			}).(TopicCreationStartedMsg)

			switch msg.AwaitCompletion().(type) {
			case TopicCreatedMsg:
			case TopicCreationErrMsg:
				t.Fatal("Unable to create topic", msg.Err)
			}

			// when
			assert.EventuallyWithT(t, func(c *assert.CollectT) {
				for i := 0; i < 55; i++ {
					now := time.Now()
					endOfToday := time.Date(
						now.Year(),
						now.Month(),
						now.Day(),
						24,
						59,
						59,
						59,
						now.Location(),
					)
					psm := ka.PublishRecord(&ProducerRecord{
						Topic:     topic,
						Key:       strconv.Itoa(i),
						Value:     []byte("{\"id\":\"123\"}"),
						Timestamp: endOfToday.Add(time.Duration(-55+i) * time.Hour),
					})

					select {
					case err := <-psm.Err:
						t.Fatal(c, "Unable to publish", err)
					case p := <-psm.Published:
						assert.True(c, p)
					}
				}
			}, 10*time.Second, 10*time.Millisecond)

			// then
			rsm := ka.ReadRecords(context.Background(), ReadDetails{
				TopicName:       topic,
				PartitionToRead: []int{0},
				StartPoint:      Today,
				Limit:           500,
			}).(*ReadingStartedMsg)

			var receivedRecords []int
			for {
				select {
				case r, ok := <-rsm.ConsumerRecord:
					if !ok {
						goto assertRecords
					}
					key, _ := strconv.Atoi(r.Key)
					receivedRecords = append(receivedRecords, key)
				}
			}

		assertRecords:
			{
				assert.Len(t, receivedRecords, 24)
			}

			// clean up
			ka.DeleteTopic(topic)
		})

		t.Run("when there are no records from previous days", func(t *testing.T) {
			topic := topicName()
			// given
			msg := ka.CreateTopic(TopicCreationDetails{
				Name:              topic,
				NumPartitions:     1,
				ReplicationFactor: 1,
			}).(TopicCreationStartedMsg)

			switch msg.AwaitCompletion().(type) {
			case TopicCreatedMsg:
			case TopicCreationErrMsg:
				t.Fatal("Unable to create topic", msg.Err)
			}

			// when
			assert.EventuallyWithT(t, func(c *assert.CollectT) {
				for i := 0; i < 55; i++ {
					twentyFourHoursAgo := time.Now().Add(-25 * time.Hour)
					psm := ka.PublishRecord(&ProducerRecord{
						Topic:     topic,
						Key:       strconv.Itoa(i),
						Value:     []byte("{\"id\":\"123\"}"),
						Timestamp: twentyFourHoursAgo.Add(time.Duration(-i) * time.Hour),
					})

					select {
					case err := <-psm.Err:
						t.Fatal(c, "Unable to publish", err)
					case p := <-psm.Published:
						assert.True(c, p)
					}
				}
			}, 10*time.Second, 10*time.Millisecond)

			// then
			rsm := ka.ReadRecords(context.Background(), ReadDetails{
				TopicName:       topic,
				PartitionToRead: []int{0},
				StartPoint:      Today,
				Limit:           500,
			}).(*ReadingStartedMsg)

			select {
			case <-rsm.NoRecordsFound:
			case <-time.After(5 * time.Second):
				t.Fatal("timed out waiting for NoRecordsFound signal")
			}

			// clean up
			ka.DeleteTopic(topic)

		})
	})

	//t.Run("Read Live", func(t *testing.T) {
	//	topic := topicName()
	//	// given
	//	msg := ka.CreateTopic(TopicCreationDetails{
	//		Name:              topic,
	//		NumPartitions:     1,
	//		ReplicationFactor: 1,
	//	}).(TopicCreationStartedMsg)
	//
	//	switch msg.AwaitCompletion().(type) {
	//	case TopicCreatedMsg:
	//	case TopicCreationErrMsg:
	//		t.Fatal("Unable to create topic", msg.Err)
	//	}
	//
	//	// existing records on topic
	//	assert.EventuallyWithT(t, func(c *assert.CollectT) {
	//		for i := 0; i < 10; i++ {
	//			psm := ka.PublishRecord(&ProducerRecord{
	//				Topic: topic,
	//				Key:   strconv.Itoa(i),
	//				Value: []byte("{\"id\":\"123\"}"),
	//			})
	//
	//			select {
	//			case err := <-psm.Err:
	//				t.Fatal(c, "Unable to publish", err)
	//			case p := <-psm.Published:
	//				assert.True(c, p)
	//			}
	//		}
	//	}, 10*time.Second, 10*time.Millisecond)
	//
	//	// when
	//	rsm := ka.ReadRecords(context.Background(), ReadDetails{
	//		TopicName:       topic,
	//		PartitionToRead: []int{0},
	//		StartPoint:      Live,
	//		Limit:           100,
	//	}).(*ReadingStartedMsg)
	//
	//	go func() {
	//		assert.EventuallyWithT(t, func(c *assert.CollectT) {
	//			for i := 10; i < 20; i++ {
	//				psm := ka.PublishRecord(&ProducerRecord{
	//					Topic: topic,
	//					Key:   strconv.Itoa(i),
	//					Value: []byte("{\"id\":\"123\"}"),
	//				})
	//
	//				select {
	//				case err := <-psm.Err:
	//					t.Fatal(c, "Unable to publish", err)
	//				case p := <-psm.Published:
	//					assert.True(c, p)
	//				}
	//			}
	//		}, 10*time.Second, 10*time.Millisecond)
	//	}()
	//
	//	var receivedRecords []int
	//	for {
	//		select {
	//		case r, ok := <-rsm.ConsumerRecord:
	//			if !ok {
	//				goto assertRecords
	//			}
	//			key, _ := strconv.Atoi(r.Key)
	//			receivedRecords = append(receivedRecords, key)
	//			if len(receivedRecords) == 10 {
	//				rsm.CancelFunc()
	//			}
	//		}
	//	}
	//
	//assertRecords:
	//	{
	//		assert.Equal(t, 19, slices.Max(receivedRecords))
	//		assert.Equal(t, 10, slices.Min(receivedRecords))
	//	}
	//
	//	// clean up
	//	ka.DeleteTopic(topic)
	//})

	t.Run("Read filtered", func(t *testing.T) {
		t.Run("with key filter", func(t *testing.T) {
			t.Run("containing", func(t *testing.T) {
				topic := topicName()
				// given
				msg := ka.CreateTopic(TopicCreationDetails{
					Name:              topic,
					NumPartitions:     1,
					ReplicationFactor: 1,
				}).(TopicCreationStartedMsg)

				switch msg.AwaitCompletion().(type) {
				case TopicCreatedMsg:
				case TopicCreationErrMsg:
					t.Fatal("Unable to create topic", msg.Err)
				}

				// when
				assert.EventuallyWithT(t, func(c *assert.CollectT) {
					for i := 0; i < 55; i++ {
						psm := ka.PublishRecord(&ProducerRecord{
							Topic: topic,
							Key:   strconv.Itoa(i),
							Value: []byte("{\"id\":\"3\"}"),
						})

						select {
						case err := <-psm.Err:
							t.Fatal(c, "Unable to publish", err)
						case p := <-psm.Published:
							assert.True(c, p)
						}
					}
				}, 10*time.Second, 10*time.Millisecond)

				// then
				rsm := ka.ReadRecords(context.Background(), ReadDetails{
					TopicName:       topic,
					PartitionToRead: []int{0},
					StartPoint:      MostRecent,
					Limit:           55,
					Filter: &Filter{
						KeySearchTerm: "1",
						KeyFilter:     ContainsFilterType,
					},
				}).(*ReadingStartedMsg)

				var receivedRecords []int
				for {
					select {
					case r, ok := <-rsm.ConsumerRecord:
						if !ok {
							goto assertRecords
						}
						key, _ := strconv.Atoi(r.Key)
						receivedRecords = append(receivedRecords, key)
						if len(receivedRecords) == 15 {
							rsm.CancelFunc()
						}
					case <-time.After(5 * time.Second):
						rsm.CancelFunc()
					}
				}

			assertRecords:
				assert.Equal(t, []int{1, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 21, 31, 41, 51}, receivedRecords)

				// clean up
				ka.DeleteTopic(topic)
			})

			t.Run("most recent with filter that matches fewer records than limit should exhaust consumption", func(t *testing.T) {
				topic := topicName()
				// given
				msg := ka.CreateTopic(TopicCreationDetails{
					Name:              topic,
					NumPartitions:     1,
					ReplicationFactor: 1,
				}).(TopicCreationStartedMsg)

				switch msg.AwaitCompletion().(type) {
				case TopicCreatedMsg:
				case TopicCreationErrMsg:
					t.Fatal("Unable to create topic", msg.Err)
				}

				// when - produce 30 messages (keys 0-29)
				assert.EventuallyWithT(t, func(c *assert.CollectT) {
					for i := 0; i < 30; i++ {
						psm := ka.PublishRecord(&ProducerRecord{
							Topic: topic,
							Key:   strconv.Itoa(i),
							Value: []byte("{\"id\":\"test\"}"),
						})

						select {
						case err := <-psm.Err:
							t.Fatal(c, "Unable to publish", err)
						case p := <-psm.Published:
							assert.True(c, p)
						}
					}
				}, 10*time.Second, 10*time.Millisecond)

				// then - consume with MostRecent and Limit=30, but filter matches only 11 records (keys containing "1")
				rsm := ka.ReadRecords(context.Background(), ReadDetails{
					TopicName:       topic,
					PartitionToRead: []int{0},
					StartPoint:      MostRecent,
					Limit:           30,
					Filter: &Filter{
						KeySearchTerm: "1",
						KeyFilter:     ContainsFilterType,
					},
				}).(*ReadingStartedMsg)

				var receivedRecords []int
				for {
					select {
					case r, ok := <-rsm.ConsumerRecord:
						if !ok {
							goto assertRecords
						}
						key, _ := strconv.Atoi(r.Key)
						receivedRecords = append(receivedRecords, key)
					case <-time.After(5 * time.Second):
						rsm.CancelFunc()
						t.Fatal("timed out waiting for consumption to end - channel should have closed")
					}
				}

			assertRecords:
				// Should have received 12 records (keys 1, 10-19, 21)
				assert.Equal(t, 12, len(receivedRecords), "expected 12 records matching filter '1'")

				// clean up
				ka.DeleteTopic(topic)
			})
		})

		t.Run("starts with", func(t *testing.T) {
			topic := topicName()
			// given
			msg := ka.CreateTopic(TopicCreationDetails{
				Name:              topic,
				NumPartitions:     1,
				ReplicationFactor: 1,
			}).(TopicCreationStartedMsg)

			switch msg.AwaitCompletion().(type) {
			case TopicCreatedMsg:
			case TopicCreationErrMsg:
				t.Fatal("Unable to create topic", msg.Err)
			}

			// when
			assert.EventuallyWithT(t, func(c *assert.CollectT) {
				for i := 0; i < 55; i++ {
					psm := ka.PublishRecord(&ProducerRecord{
						Topic: topic,
						Key:   strconv.Itoa(i),
						Value: []byte("{\"id\":\"3\"}"),
					})

					select {
					case err := <-psm.Err:
						t.Fatal(c, "Unable to publish", err)
					case p := <-psm.Published:
						assert.True(c, p)
					}
				}
			}, 10*time.Second, 10*time.Millisecond)

			// then
			rsm := ka.ReadRecords(context.Background(), ReadDetails{
				TopicName:       topic,
				PartitionToRead: []int{0},
				StartPoint:      MostRecent,
				Limit:           55,
				Filter: &Filter{
					KeySearchTerm: "1",
					KeyFilter:     StartsWithFilterType,
				},
			}).(*ReadingStartedMsg)

			var receivedRecords []int
			for {
				select {
				case r, ok := <-rsm.ConsumerRecord:
					if !ok {
						goto assertRecords
					}
					key, _ := strconv.Atoi(r.Key)
					receivedRecords = append(receivedRecords, key)
					if len(receivedRecords) == 11 {
						rsm.CancelFunc()
					}
				case <-time.After(5 * time.Second):
					rsm.CancelFunc()
				}
			}

		assertRecords:
			assert.Equal(t, []int{1, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}, receivedRecords)

			// clean up
			ka.DeleteTopic(topic)
		})

	})

	t.Run("Transactional topic where last offset is a control record", func(t *testing.T) {
		topic := topicName()
		createSinglePartitionTopic(t, topic)
		produceTransactionally(t, topic, 10)

		for _, sp := range []StartPoint{Beginning, MostRecent} {
			// when
			rsm := ka.ReadRecords(context.Background(), ReadDetails{
				TopicName:       topic,
				PartitionToRead: []int{0},
				StartPoint:      sp,
				Limit:           50,
				Filter:          &Filter{},
			}).(*ReadingStartedMsg)

			// then
			started := time.Now()
			records := consumeUntilEnded(t, rsm, 10*time.Second)
			assert.Len(t, records, 10)
			assert.Less(t, time.Since(started), idleConsumptionTimeout/2,
				"consumption should end on the last delivered record without waiting")
		}

		// clean up
		ka.DeleteTopic(topic)
	})

	t.Run("Read from beginning with filter not matching the last record", func(t *testing.T) {
		topic := topicName()
		createSinglePartitionTopic(t, topic)

		for i := 0; i < 5; i++ {
			psm := ka.PublishRecord(&ProducerRecord{
				Topic: topic,
				Key:   strconv.Itoa(i),
				Value: []byte("value"),
			})
			select {
			case err := <-psm.Err:
				t.Fatal("Unable to publish", err)
			case <-psm.Published:
			}
		}

		// when
		rsm := ka.ReadRecords(context.Background(), ReadDetails{
			TopicName:       topic,
			PartitionToRead: []int{0},
			StartPoint:      Beginning,
			Limit:           50,
			Filter: &Filter{
				KeyFilter:     ContainsFilterType,
				KeySearchTerm: "1",
			},
		}).(*ReadingStartedMsg)

		// then
		records := consumeUntilEnded(t, rsm, 10*time.Second)
		assert.Len(t, records, 1)

		// clean up
		ka.DeleteTopic(topic)
	})
}

func createSinglePartitionTopic(t *testing.T, topic string) {
	msg := ka.CreateTopic(TopicCreationDetails{
		Name:              topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}).(TopicCreationStartedMsg)

	switch msg.AwaitCompletion().(type) {
	case TopicCreatedMsg:
	case TopicCreationErrMsg:
		t.Fatal("Unable to create topic", msg.Err)
	}
}

func TestLastDataOffset(t *testing.T) {
	sk := ka.(*SaramaKafkaAdmin)

	t.Run("skips trailing commit marker", func(t *testing.T) {
		topic := topicName()
		createSinglePartitionTopic(t, topic)
		produceTransactionally(t, topic, 10)

		end, err := sk.client.GetOffset(topic, 0, sarama.OffsetNewest)
		assert.NoError(t, err)
		assert.Equal(t, int64(11), end)

		last, err := sk.lastDataOffset(topic, 0, 0, end)
		assert.NoError(t, err)
		assert.Equal(t, int64(9), last)

		ka.DeleteTopic(topic)
	})

	t.Run("includes aborted records when reading uncommitted", func(t *testing.T) {
		topic := topicName()
		createSinglePartitionTopic(t, topic)
		produceTransactionally(t, topic, 5)
		produceAbortedTransaction(t, topic, 3)

		// 5 records + commit marker + 3 records + abort marker
		var end int64
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			var err error
			end, err = sk.client.GetOffset(topic, 0, sarama.OffsetNewest)
			assert.NoError(c, err)
			assert.Equal(c, int64(10), end)
		}, 5*time.Second, 50*time.Millisecond)

		last, err := sk.lastDataOffset(topic, 0, 0, end)
		assert.NoError(t, err)
		assert.Equal(t, int64(8), last)

		ka.DeleteTopic(topic)
	})

	t.Run("no delivered records in range", func(t *testing.T) {
		topic := topicName()
		createSinglePartitionTopic(t, topic)
		produceTransactionally(t, topic, 2)

		// only the commit marker at offset 2
		last, err := sk.lastDataOffset(topic, 0, 2, 3)
		assert.NoError(t, err)
		assert.Equal(t, int64(-1), last)

		ka.DeleteTopic(topic)
	})
}

func produceTransactionally(t *testing.T, topic string, count int) {
	producer := newTransactionalProducer(t, topic)
	defer producer.Close()

	sendInTransaction(t, producer, topic, count)
	if err := producer.CommitTxn(); err != nil {
		t.Fatal("Unable to commit transaction", err)
	}
}

func produceAbortedTransaction(t *testing.T, topic string, count int) {
	producer := newTransactionalProducer(t, topic)
	defer producer.Close()

	sendInTransaction(t, producer, topic, count)
	if err := producer.AbortTxn(); err != nil {
		t.Fatal("Unable to abort transaction", err)
	}
}

func newTransactionalProducer(t *testing.T, topic string) sarama.SyncProducer {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Idempotent = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Transaction.ID = "txn-" + topic
	cfg.Net.MaxOpenRequests = 1

	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		t.Fatal("Unable to create transactional producer", err)
	}
	return producer
}

func sendInTransaction(t *testing.T, producer sarama.SyncProducer, topic string, count int) {
	if err := producer.BeginTxn(); err != nil {
		t.Fatal("Unable to begin transaction", err)
	}
	for i := 0; i < count; i++ {
		_, _, err := producer.SendMessage(&sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(strconv.Itoa(i)),
			Value: sarama.StringEncoder("value"),
		})
		if err != nil {
			t.Fatal("Unable to send message", err)
		}
	}
}

// consumeUntilEnded collects records until the record channel is closed,
// failing the test when consumption does not end within the timeout.
func consumeUntilEnded(t *testing.T, rsm *ReadingStartedMsg, timeout time.Duration) []ConsumerRecord {
	var records []ConsumerRecord
	deadline := time.After(timeout)
	for {
		select {
		case r, ok := <-rsm.ConsumerRecord:
			if !ok {
				return records
			}
			records = append(records, r)
		case err, ok := <-rsm.Err:
			if !ok {
				// closed on shutdown, keep draining the records
				rsm.Err = nil
				continue
			}
			rsm.CancelFunc()
			t.Fatal("unexpected error during consumption", err)
		case <-deadline:
			rsm.CancelFunc()
			t.Fatal("timed out waiting for consumption to end")
		}
	}
}

type want struct {
	start int64
	end   int64
}

type determineStartingOffsetTest struct {
	name        string
	want        want
	readDetails ReadDetails
	offsets     offsets
}

func TestDetermineStartingOffset(t *testing.T) {
	var tests = []determineStartingOffsetTest{
		{
			name: "beginning one partition enough records available",
			readDetails: ReadDetails{
				TopicName:       "test-topic",
				PartitionToRead: []int{0},
				StartPoint:      Beginning,
				Limit:           50,
			},
			offsets: offsets{
				start: 1,
				end:   291,
			},

			want: want{
				start: 1,
				end:   50,
			},
		},
		{
			name: "beginning multiple partition enough records available",
			readDetails: ReadDetails{
				TopicName:       "test-topic",
				PartitionToRead: []int{0, 1, 2, 3, 4},
				StartPoint:      Beginning,
				Limit:           50,
			},
			offsets: offsets{
				start: 1,
				end:   291,
			},

			want: want{
				start: 1,
				end:   10,
			},
		},
		{
			name: "beginning one partition not enough records available",
			readDetails: ReadDetails{
				TopicName:       "test-topic",
				PartitionToRead: []int{0},
				StartPoint:      Beginning,
				Limit:           50,
			},
			offsets: offsets{
				start: 55,
				end:   76,
			},
			want: want{
				start: 55,
				end:   75,
			},
		},
		{
			name: "beginning multiple partition not enough records available",
			readDetails: ReadDetails{
				TopicName:       "test-topic",
				PartitionToRead: []int{0, 2, 3, 4},
				StartPoint:      Beginning,
				Limit:           50,
			},
			offsets: offsets{
				start: 10,
				end:   12,
			},

			want: want{
				start: 10,
				end:   11,
			},
		},
		{
			name: "most recent one partition enough records available",
			readDetails: ReadDetails{
				TopicName:       "test-topic",
				PartitionToRead: []int{0},
				StartPoint:      MostRecent,
				Limit:           50,
			},
			offsets: offsets{
				start: 1,
				end:   291,
			},

			want: want{
				start: 240,
				end:   290,
			},
		},
		{
			name: "most recent one partition not enough records available",
			readDetails: ReadDetails{
				TopicName:       "test-topic",
				PartitionToRead: []int{0},
				StartPoint:      MostRecent,
				Limit:           50,
			},
			offsets: offsets{
				start: 278,
				end:   291,
			},

			want: want{
				start: 278,
				end:   290,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			offset := ka.(*SaramaKafkaAdmin).determineReadingOffsets(
				test.readDetails,
				test.offsets,
			)
			assert.Equal(t, test.want.start, offset.start, "unexpected start")
			assert.Equal(t, test.want.end, offset.end, "unexpected end")
		})
	}
}

func TestConsumerRecordPayloadType(t *testing.T) {
	t.Run("returns Avro when schema is present", func(t *testing.T) {
		record := &ConsumerRecord{
			Payload: serdes.DesData{
				Schema: `{"type": "record", "name": "Test"}`,
				Value:  `{"field": "value"}`,
			},
		}

		assert.Equal(t, "Avro", record.PayloadType())
	})

	t.Run("returns Plain Text when value is empty", func(t *testing.T) {
		record := &ConsumerRecord{
			Payload: serdes.DesData{
				Schema: "",
				Value:  "",
			},
		}

		assert.Equal(t, "Plain Text", record.PayloadType())
	})

	t.Run("returns Plain Text when value is whitespace only", func(t *testing.T) {
		record := &ConsumerRecord{
			Payload: serdes.DesData{
				Schema: "",
				Value:  "   \n\t  ",
			},
		}

		assert.Equal(t, "Plain Text", record.PayloadType())
	})

	t.Run("returns Plain Json when value is valid JSON", func(t *testing.T) {
		record := &ConsumerRecord{
			Payload: serdes.DesData{
				Schema: "",
				Value:  `{"name": "John", "age": 30}`,
			},
		}

		assert.Equal(t, "Plain Json", record.PayloadType())
	})

	t.Run("returns Plain XML when value is valid XML", func(t *testing.T) {
		record := &ConsumerRecord{
			Payload: serdes.DesData{
				Schema: "",
				Value:  `<root><name>John</name><age>30</age></root>`,
			},
		}

		assert.Equal(t, "Plain XML", record.PayloadType())
	})

	t.Run("returns Plain Text for invalid JSON and XML", func(t *testing.T) {
		record := &ConsumerRecord{
			Payload: serdes.DesData{
				Schema: "",
				Value:  "This is just plain text",
			},
		}

		assert.Equal(t, "Plain Text", record.PayloadType())
	})

	t.Run("returns Plain Text for malformed JSON", func(t *testing.T) {
		record := &ConsumerRecord{
			Payload: serdes.DesData{
				Schema: "",
				Value:  `{"name": "John", "age":`,
			},
		}

		assert.Equal(t, "Plain Text", record.PayloadType())
	})

	t.Run("returns Plain Text for malformed XML", func(t *testing.T) {
		record := &ConsumerRecord{
			Payload: serdes.DesData{
				Schema: "",
				Value:  `<root><name>John</name>`,
			},
		}

		assert.Equal(t, "Plain Text", record.PayloadType())
	})
}

func TestAwaitRecordsEndsAfterShutdown(t *testing.T) {
	for i := 0; i < 200; i++ {
		_, cancel := context.WithCancel(context.Background())
		rsm := &ReadingStartedMsg{
			ConsumerRecord: make(chan ConsumerRecord, 3),
			Err:            make(chan error, 1),
			EmptyTopic:     make(chan bool, 1),
			NoRecordsFound: make(chan bool, 1),
			CancelFunc:     cancel,
		}
		for o := 0; o < 3; o++ {
			rsm.ConsumerRecord <- ConsumerRecord{Offset: int64(o)}
		}
		rsm.shutdown()

		var received int
		next := rsm.AwaitRecord
		for {
			msg := next()
			if r, ok := msg.(ConsumerRecordReceived); ok {
				received += len(r.Records)
				next = r.AwaitNextRecord
				continue
			}
			if _, ok := msg.(ConsumptionEndedMsg); !ok {
				t.Fatalf("expected ConsumptionEndedMsg, got %T (%v)", msg, msg)
			}
			break
		}
		assert.Equal(t, 3, received)
	}
}
