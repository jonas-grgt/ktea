package kadmin

import (
	"fmt"
	"ktea/config"
	"ktea/sradmin"
	"sort"
	"time"

	"github.com/IBM/sarama"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

type SaramaKafkaAdmin struct {
	client   sarama.Client
	admin    sarama.ClusterAdmin
	addrs    []string
	config   *sarama.Config
	producer sarama.SyncProducer
	sra      sradmin.SrAdmin
}

func (s *SaramaKafkaAdmin) GetClusterConfig() (ClusterConfig, error) {
	MaybeIntroduceLatency()
	saramaBrokers, _, err := s.admin.DescribeCluster()
	if err != nil {
		log.Error("Failed to describe cluster", "error", err)
		return ClusterConfig{}, err
	}
	sort.Slice(saramaBrokers, func(i, j int) bool {
		return saramaBrokers[i].ID() < saramaBrokers[j].ID()
	})

	brokers := make([]BrokerConfig, 0)
	for _, saramaBroker := range saramaBrokers {
		log.Info("Fetching config for broker", "brokerID", saramaBroker.ID(), "addr", saramaBroker.Addr())
		resource := sarama.ConfigResource{
			Type: sarama.BrokerResource,
			Name: fmt.Sprintf("%d", saramaBroker.ID()),
		}

		entries, err := s.admin.DescribeConfig(resource)
		if err != nil {
			log.Printf("Failed to describe config for broker %d: %v", saramaBroker.ID(), err)
			continue
		}

		configMap := make(map[string]string)
		for _, entry := range entries {
			configMap[entry.Name] = entry.Value
		}

		brokers = append(brokers, BrokerConfig{
			ID:      saramaBroker.ID(),
			Addr:    saramaBroker.Addr(),
			Configs: configMap,
		})

	}

	return ClusterConfig{
		Brokers: brokers,
	}, nil
}

type ConnCheckStartedMsg struct {
	Cluster   *config.Cluster
	Connected chan bool
	Err       chan error
}

func (c *ConnCheckStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-c.Connected:
		return ConnCheckSucceededMsg{}
	case err := <-c.Err:
		return ConnCheckErrMsg{Err: err}
	}
}

type ConnCheckSucceededMsg struct{}

type ConnCheckErrMsg struct {
	Err error
}

func ToConnectionDetails(cluster *config.Cluster) ConnectionDetails {
	var saslConfig *SASLConfig
	if cluster.SASLConfig != nil {
		var protocol SASLProtocol
		switch cluster.SASLConfig.SecurityProtocol {
		// SSL, to make wrongly configured PLAINTEXT protocols (as SSL) compatible. Should be removed in the future.
		case config.SASLPlaintextSecurityProtocol, "SSL":
			protocol = PlainText
		default:
			panic(fmt.Sprintf("Unknown SASL protocol: %s", cluster.SASLConfig.SecurityProtocol))
		}

		saslConfig = &SASLConfig{
			Username: cluster.SASLConfig.Username,
			Password: cluster.SASLConfig.Password,
			Protocol: protocol,
		}
	}

	connDetails := ConnectionDetails{
		BootstrapServers: cluster.BootstrapServers,
		SASLConfig:       saslConfig,
		SSLEnabled:       cluster.SSLEnabled,
	}
	return connDetails
}

func NewSaramaKadmin(cd ConnectionDetails) (Kadmin, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	cfg.Net.TLS.Enable = cd.SSLEnabled

	if cd.SASLConfig != nil {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		cfg.Net.SASL.User = cd.SASLConfig.Username
		cfg.Net.SASL.Password = cd.SASLConfig.Password
	}

	client, err := sarama.NewClient(cd.BootstrapServers, cfg)
	if err != nil {
		return nil, err
	}

	admin, err := sarama.NewClusterAdmin(cd.BootstrapServers, cfg)
	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}

	return &SaramaKafkaAdmin{
		client:   client,
		admin:    admin,
		addrs:    cd.BootstrapServers,
		producer: producer,
		config:   cfg,
	}, nil
}

func SaramaConnectivityChecker(cluster *config.Cluster) tea.Msg {
	connectedChan := make(chan bool)
	errChan := make(chan error)

	cd := ToConnectionDetails(cluster)
	cfg := sarama.NewConfig()

	cfg.Net.TLS.Enable = cd.SSLEnabled

	if cd.SASLConfig != nil {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		cfg.Net.SASL.User = cd.SASLConfig.Username
		cfg.Net.SASL.Password = cd.SASLConfig.Password
		cfg.Net.DialTimeout = 5 * time.Second
		cfg.Net.ReadTimeout = 5 * time.Second
		cfg.Net.WriteTimeout = 5 * time.Second
	}

	go doCheckConnectivity(cd, cfg, errChan, connectedChan)

	return ConnCheckStartedMsg{
		Cluster:   cluster,
		Connected: connectedChan,
		Err:       errChan,
	}
}

func doCheckConnectivity(cd ConnectionDetails, config *sarama.Config, errChan chan error, connectedChan chan bool) {
	MaybeIntroduceLatency()
	c, err := sarama.NewClient(cd.BootstrapServers, config)
	if err != nil {
		errChan <- err
		return
	}
	defer func(c sarama.Client) {
		err := c.Close()
		if err != nil {
			log.Error("Unable to close connectivity check connection", err)
		}
	}(c)
	connectedChan <- true
}
