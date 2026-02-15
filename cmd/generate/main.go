package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"ktea/config"
	"ktea/kadmin"
	"ktea/sradmin"
	"ktea/styles"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/linkedin/goavro/v2"
	"golang.org/x/exp/rand"
)

type eventGenFunc func(id string) (string, interface{})

type topic struct {
	name      string
	compacted bool
}
type generationData struct {
	topic
	subject string
	schema  []string
	eventGenFunc
}

func main() {
	genData := []generationData{
		{topic{"dev.finance.invoice", false},
			"dev.finance.invoice-io.jonasg.ktea.invoice.InvoiceCreated",
			[]string{`
			{
				"type": "record",
				"name": "InvoiceCreated",
				"namespace": "io.jonasg.ktea.invoice",
				"doc": "Schema for the InvoiceCreated event.",
				"fields": [
					{"name": "id", "type": "string", "doc": "Unique identifier for the invoice."},
					{"name": "customerId", "type": "string", "doc": "Unique identifier for the customer."},
					{"name": "amount", "type": "bytes", "logicalType": "decimal", "precision": 4, "scale": 2, "doc": "Total amount of the invoice."},
					{"name": "currency", "type": "string", "doc": "Currency of the invoice amount."},
					{"name": "issueDate", "type": "string", "doc": "Date when the invoice was issued, in ISO 8601 format."},
					{"name": "dueDate", "type": "string", "doc": "Date when the invoice is due, in ISO 8601 format."},
					{"name": "status", "type": "string", "doc": "Current status of the invoice (e.g., 'Paid', 'Pending')."},
					{"name": "description", "type": "string", "doc": "Description or notes about the invoice."}
				]
			}
		`},
			func(id string) (string, interface{}) {
				return id, map[string]interface{}{
					"id":          id,
					"customerId":  uuid.New().String(),
					"amount":      big.NewRat(100, 100),
					"currency":    "USD",
					"issueDate":   time.Now().Format(time.RFC3339),
					"dueDate":     time.Now().AddDate(0, 0, 30).Format(time.RFC3339),
					"status":      "Pending",
					"description": "Invoice for services rendered.",
				}
			},
		},

		{
			topic{"dev.finance.payment", false},
			"dev.finance.payment-io.jonasg.ktea.payment.PaymentProcessed",
			[]string{`
	{
		"type": "record",
		"name": "PaymentProcessed",
		"namespace": "io.jonasg.ktea.payment",
		"doc": "Schema for the PaymentProcessed event.",
		"fields": [
			{"name": "id", "type": "string", "doc": "Unique identifier for the payment."},
			{"name": "invoiceId", "type": "string", "doc": "Unique identifier for the associated invoice."},
			{"name": "amount", "type": "bytes", "logicalType": "decimal", "precision": 4, "scale": 2, "doc": "Amount of the payment."},
			{"name": "currency", "type": "string", "doc": "Currency of the payment amount."},
			{"name": "paymentDate", "type": "string", "doc": "Date when the payment was made, in ISO 8601 format."},
			{"name": "status", "type": "string", "doc": "Current status of the payment (e.g., 'Completed', 'Failed')."},
			{"name": "method", "type": "string", "doc": "Payment method used (e.g., 'Credit Card', 'Bank Transfer')."}
		]
	}
	`},
			func(id string) (string, interface{}) {
				return id, map[string]interface{}{
					"id":          id,
					"invoiceId":   uuid.New().String(),
					"amount":      big.NewRat(500, 100),
					"currency":    "USD",
					"paymentDate": time.Now().Format(time.RFC3339),
					"status":      "Completed",
					"method":      "Credit Card",
				}
			},
		},

		{
			topic{"dev.order.checkout", false},
			"dev.order.checkout-io.jonasg.ktea.order.CheckoutInitiated",
			[]string{`
	{
		"type": "record",
		"name": "CheckoutInitiated",
		"namespace": "io.jonasg.ktea.order",
		"doc": "Schema for the CheckoutInitiated event.",
		"fields": [
			{"name": "id", "type": "string", "doc": "Unique identifier for the checkout."},
			{"name": "cartId", "type": "string", "doc": "Unique identifier for the cart."},
			{"name": "totalAmount", "type": "bytes", "logicalType": "decimal", "precision": 10, "scale": 2, "doc": "Total amount for the checkout."},
			{"name": "currency", "type": "string", "doc": "Currency of the total amount."},
			{"name": "checkoutDate", "type": "string", "doc": "Date when the checkout was initiated, in ISO 8601 format."}
		]
	}
	`},
			func(id string) (string, interface{}) {
				return id, map[string]interface{}{
					"id":           id,
					"cartId":       uuid.New().String(),
					"totalAmount":  big.NewRat(1500, 100),
					"currency":     "USD",
					"checkoutDate": time.Now().Format(time.RFC3339),
				}
			},
		},
		{
			topic{"dev.order.shipment", false},
			"dev.order.shipment-io.jonasg.ktea.order.ShipmentCreated",
			[]string{`
	{
		"type": "record",
		"name": "ShipmentCreated",
		"namespace": "io.jonasg.ktea.order",
		"doc": "Schema for the ShipmentCreated event.",
		"fields": [
			{"name": "id", "type": "string", "doc": "Unique identifier for the shipment."},
			{"name": "orderId", "type": "string", "doc": "Unique identifier for the order."},
			{"name": "shipmentDate", "type": "string", "doc": "Date when the shipment was created, in ISO 8601 format."},
			{"name": "carrier", "type": "string", "doc": "Carrier responsible for the shipment."},
			{"name": "trackingNumber", "type": "string", "doc": "Tracking number for the shipment."}
		]
	}
	`},
			func(id string) (string, interface{}) {
				return id, map[string]interface{}{
					"id":             id,
					"orderId":        uuid.New().String(),
					"shipmentDate":   time.Now().Format(time.RFC3339),
					"carrier":        "FedEx",
					"trackingNumber": "123456789",
				}
			},
		},
		{
			topic{"dev.product.stock", false},
			"dev.product.stock-io.jonasg.ktea.product.StockUpdated",
			[]string{`
	{
		"type": "record",
		"name": "StockUpdated",
		"namespace": "io.jonasg.ktea.product",
		"doc": "Schema for the StockUpdated event.",
		"fields": [
			{"name": "id", "type": "string", "doc": "Unique identifier for the stock update."},
			{"name": "productId", "type": "string", "doc": "Unique identifier for the product."},
			{"name": "quantity", "type": "int", "doc": "Quantity of the product in stock."},
			{"name": "updateDate", "type": "string", "doc": "Date when the stock was updated, in ISO 8601 format."}
		]
	}
	`},
			func(id string) (string, interface{}) {
				return id, map[string]interface{}{
					"id":         id,
					"productId":  uuid.New().String(),
					"quantity":   100,
					"updateDate": time.Now().Format(time.RFC3339),
				}
			},
		},
		{
			topic{"dev.product.category", false},
			"dev.product.category-io.jonasg.ktea.product.CategoryAssigned",
			[]string{`
	{
		"type": "record",
		"name": "CategoryAssigned",
		"namespace": "io.jonasg.ktea.product",
		"doc": "Schema for the CategoryAssigned event.",
		"fields": [
			{"name": "id", "type": "string", "doc": "Unique identifier for the category assignment."},
			{"name": "productId", "type": "string", "doc": "Unique identifier for the product."},
			{"name": "category", "type": "string", "doc": "Category assigned to the product."},
			{"name": "assignmentDate", "type": "string", "doc": "Date when the category was assigned, in ISO 8601 format."}
		]
	}
	`},
			func(id string) (string, interface{}) {
				return id, map[string]interface{}{
					"id":             id,
					"productId":      uuid.New().String(),
					"category":       "Electronics",
					"assignmentDate": time.Now().Format(time.RFC3339),
				}
			},
		},
		{
			topic{"dev.product.price", false},
			"dev.product.price-io.jonasg.ktea.product.PriceUpdated",
			[]string{`
	{
		"type": "record",
		"name": "PriceUpdated",
		"namespace": "io.jonasg.ktea.product",
		"doc": "Schema for the PriceUpdated event.",
		"fields": [
			{"name": "id", "type": "string", "doc": "Unique identifier for the price update."},
			{"name": "productId", "type": "string", "doc": "Unique identifier for the product."},
			{"name": "price", "type": "bytes", "logicalType": "decimal", "precision": 10, "scale": 2, "doc": "Updated price of the product."},
			{"name": "currency", "type": "string", "doc": "Currency of the price."},
			{"name": "updateDate", "type": "string", "doc": "Date when the price was updated, in ISO 8601 format."}
		]
	}
	`},
			func(id string) (string, interface{}) {
				return id, map[string]interface{}{
					"id":         id,
					"productId":  uuid.New().String(),
					"price":      big.NewRat(2000, 100),
					"currency":   "USD",
					"updateDate": time.Now().Format(time.RFC3339),
				}
			},
		},
		{
			topic{"dev.customer.profile", true},
			"dev.customer.profile-io.jonasg.ktea.customer.ProfileUpdated",
			[]string{`
	{
		"type": "record",
		"name": "ProfileUpdated",
		"namespace": "io.jonasg.ktea.customer",
		"doc": "Schema for the ProfileUpdated event.",
		"fields": [
			{"name": "id", "type": "string", "doc": "Unique identifier for the profile update."},
			{"name": "customerId", "type": "string", "doc": "Unique identifier for the customer."},
			{"name": "updateDate", "type": "string", "doc": "Date when the profile was updated, in ISO 8601 format."},
			{"name": "changes", "type": "string", "doc": "Description of the changes made to the profile."}
		]
	}
	`},
			func(_ string) (string, interface{}) {
				id := fmt.Sprintf("%d", rand.Intn(10)+1)
				return id, map[string]interface{}{
					"id":         id,
					"customerId": uuid.New().String(),
					"updateDate": time.Now().Format(time.RFC3339),
					"changes":    "Updated email address and phone number.",
				}
			},
		},
		{
			topic{"dev.customer.action", false},
			"dev.customer.action-io.jonasg.ktea.customer.ActionLogged",
			[]string{
				`
	{
		"type": "record",
		"name": "ActionLogged",
		"namespace": "io.jonasg.ktea.customer",
		"doc": "Schema for the ActionLogged event.",
		"fields": [
			{"name": "id", "type": "string", "doc": "Unique identifier for the action log."},
			{"name": "customerId", "type": "string", "doc": "Unique identifier for the customer."},
			{"name": "action", "type": "string", "doc": "Description of the action performed by the customer."},
			{"name": "actionDate", "type": "string", "doc": "Date when the action was performed, in ISO 8601 format."}
		]
	}
	`},
			func(id string) (string, interface{}) {
				return id, map[string]interface{}{
					"id":         id,
					"customerId": uuid.New().String(),
					"action":     "Logged in to the system.",
					"actionDate": time.Now().Format(time.RFC3339),
				}
			},
		},
		{
			topic{"dev.customer.feedback", false},
			"dev.customer.feedback-io.jonasg.ktea.customer.FeedbackReceived",
			[]string{`
	{
		"type": "record",
		"name": "FeedbackReceived",
		"namespace": "io.jonasg.ktea.customer",
		"doc": "Schema for the FeedbackReceived event.",
		"fields": [
			{"name": "id", "type": "string", "doc": "Unique identifier for the feedback."},
			{"name": "customerId", "type": "string", "doc": "Unique identifier for the customer."},
			{"name": "feedback", "type": "string", "doc": "Content of the feedback provided by the customer."},
			{"name": "feedbackDate", "type": "string", "doc": "Date when the feedback was provided, in ISO 8601 format."}
		]
	}
	`},
			func(id string) (string, interface{}) {
				return id, map[string]interface{}{
					"id":           id,
					"customerId":   uuid.New().String(),
					"feedback":     "Great service!",
					"feedbackDate": time.Now().Format(time.RFC3339),
				}
			},
		},
	}

	ka, sa := getAdmins()

	wg := sync.WaitGroup{}

	for _, gd := range genData {
		wg.Go(func() {
			generateData(ka, gd, sa, 1000)
		})
	}

	wg.Wait()

	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	brokers := []string{"localhost:9092"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	groupHandler := &handler{
		cancel: cancel,
		limit:  2000,
	}
	if group, err := sarama.NewConsumerGroup(brokers, "generator", cfg); err != nil {
		panic(fmt.Sprintf("Failed to create consumer group: %v", err))
	} else {
		defer func() { _ = group.Close() }()
		for {
			if ctx.Err() != nil {
				break
			}
			if err := group.Consume(ctx, []string{"dev.finance.invoice", "dev.finance.payment"}, groupHandler); err != nil {
				fmt.Printf("Error from consumer: %v", err)
				cancel()
			}
		}
	}

}

func generateData(ka kadmin.Kadmin, gd generationData, sa sradmin.Client, count int) {
	if !topicExists(ka, gd.topic.name) {
		createTopic(ka, gd.topic)
	}

	if !subjectExists(sa, gd.subject) {
		for _, s := range gd.schema {
			registerSchema(sa, gd.subject, s)
		}
	}

	schemaInfo := getLatestSchema(sa, gd.subject)

	for i := 0; i < count; i++ {
		id := uuid.New().String()
		id, event := gd.eventGenFunc(id)
		publish(ka, gd.topic.name, id, event, schemaInfo)
	}
	fmt.Printf("Published 1000 events to topic %s with subject %s\n", gd.topic.name, gd.subject)
}

type handler struct {
	count  int
	limit  int
	cancel context.CancelFunc
}

func (h *handler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *handler) ConsumeClaim(s sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.count++
		fmt.Println(fmt.Sprintf("Consumed message number %d", h.count))
		s.MarkMessage(msg, "")
		if h.count >= h.limit {
			h.cancel()
			return nil
		}
	}
	return nil
}

func publish(ka kadmin.Kadmin, topic string, id string, event interface{}, schemaInfo sradmin.Schema) {
	//personJson, _ := json.Marshal(event)
	codec, _ := goavro.NewCodec(schemaInfo.Value)
	valueBytes, err := codec.BinaryFromNative(nil, event)
	if err != nil {
		panic(fmt.Sprintf("Failed to convert JSON to native Avro: %v", err))
	}
	//valueBytes, _ := codec.BinaryFromNative(nil, native)
	schemaId, err := strconv.Atoi(schemaInfo.Id)
	if err != nil {
		panic(fmt.Sprintf("Failed to convert schema ID to bytes: %v", err))
	}
	schemaIDBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(schemaIDBytes, uint32(schemaId))

	var record []byte
	record = append(record, byte(0))
	record = append(record, schemaIDBytes...)
	record = append(record, valueBytes...)

	msg := ka.PublishRecord(&kadmin.ProducerRecord{
		Key:       id,
		Value:     record,
		Topic:     topic,
		Partition: nil,
		Headers: map[string]string{
			"content-type": "application/vnd.apache.avro+json",
			"eventId":      id,
			"eventType":    "ProductCreated",
			"eventSource":  "ktea",
			"eventVersion": "1.0",
			"eventTime":    time.Now().String(),
		},
	})
	switch msg := msg.AwaitCompletion().(type) {
	case kadmin.PublicationSucceeded:
	case kadmin.PublicationFailed:
		panic(fmt.Sprintf("Failed to publish message %v", msg.Err))
	}
}

func getLatestSchema(sa sradmin.Client, subject string) sradmin.Schema {
	msg := sa.GetLatestSchemaBySubject(subject).(sradmin.FetchingLatestSchemaBySubjectMsg)

	var schemaInfo sradmin.Schema
	switch msg := msg.AwaitCompletion().(type) {
	case sradmin.LatestSchemaBySubjectReceived:
		fmt.Println("Latest schema fetched successfully for subject:", subject)
		schemaInfo = msg.Schema
	case sradmin.FailedToFetchLatestSchemaBySubject:
		panic(fmt.Sprintf("Failed to get latest schema by subject: %v", msg.Err))
	}
	return schemaInfo
}

func registerSchema(srAdmin sradmin.Client, subject string, schema string) {
	msg := srAdmin.CreateSchema(sradmin.SubjectCreationDetails{
		Subject: subject,
		Schema:  schema,
	}).(sradmin.SchemaCreationStartedMsg)

	switch msg := msg.AwaitCompletion().(type) {
	case sradmin.SchemaCreatedMsg:
		fmt.Println("Schema created successfully for subject:", subject)
	case sradmin.SchemaCreationErrMsg:
		panic(fmt.Sprintf("Failed to create schema for subject %s: %v", subject, msg.Err))
	}
}

func getAdmins() (kadmin.Kadmin, sradmin.Client) {
	ka, err := kadmin.NewSaramaKadmin(&config.Cluster{
		Name:             "generate",
		Color:            styles.ColorGreen,
		Active:           true,
		BootstrapServers: []string{"localhost:9092"},
		SASLConfig: config.SASLConfig{
			AuthMethod: config.AuthMethodNone,
		},
		SchemaRegistry: nil,
		TLSConfig: config.TLSConfig{
			Enable: false,
		},
		KafkaConnectClusters: nil,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create Kafka admin client: %v", err))
	}

	sa := sradmin.New(&config.SchemaRegistryConfig{
		Url:      "http://localhost:8081",
		Username: "",
		Password: "",
	})

	return ka, sa
}

func createTopic(ka kadmin.Kadmin, topic topic) {
	partitions := 1
	if (rand.Intn(1) % 2) == 0 {
		partitions = rand.Intn(10) + 1
	}
	var properties = map[string]string{}
	if topic.compacted {
		properties["cleanup.policy"] = "compact"
	}
	tm := ka.CreateTopic(kadmin.TopicCreationDetails{
		Name:              topic.name,
		NumPartitions:     partitions,
		ReplicationFactor: 1,
		Properties:        properties,
	}).(kadmin.TopicCreationStartedMsg)

	switch msg := tm.AwaitCompletion().(type) {
	case kadmin.TopicCreatedMsg:
		fmt.Printf("Topic %s created successfully", topic.name)
	case kadmin.TopicCreationErrMsg:
		panic(fmt.Sprintf("Failed to create topic: %v", msg.Err))
	}
}

func topicExists(ka kadmin.Kadmin, expectedTopic string) bool {
	msg := ka.ListTopics().(kadmin.TopicListingStartedMsg)
	switch msg := msg.AwaitTopicListCompletion().(type) {
	case kadmin.TopicsListedMsg:
		topics := msg.Topics
		for _, topic := range topics {
			if topic.Name == expectedTopic {
				fmt.Println("Topic " + expectedTopic + " already exists")
				return true
			}
		}
	case kadmin.TopicListedErrorMsg:
		panic(fmt.Sprintf("Failed to list topics: %v", msg.Err))
	}
	return false
}

func subjectExists(srAdmin sradmin.Client, subject string) bool {
	msg := srAdmin.ListSubjects().(sradmin.SubjectListingStartedMsg)
	switch msg := msg.AwaitCompletion().(type) {
	case sradmin.SubjectsListedMsg:
		for _, s := range msg.Subjects {
			if s.Name == subject {
				return true
			}
		}
	case sradmin.SubjectListingErrorMsg:
		panic(fmt.Sprintf("Failed to list subjects: %v", msg.Err))
	}
	return false
}
