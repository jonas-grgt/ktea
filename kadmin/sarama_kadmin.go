package kadmin

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"ktea/config"
	"ktea/sradmin"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/burdiyan/kafkautil"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

type SaramaKafkaAdmin struct {
	client   sarama.Client
	admin    sarama.ClusterAdmin
	addrs    []string
	config   *sarama.Config
	producer sarama.SyncProducer
	sra      sradmin.Client
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

func ToSaramaCfg(cluster *config.Cluster) *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Partitioner = kafkautil.NewJVMCompatiblePartitioner
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	if cluster.TLSConfig.Enable {

		var caCertPool *x509.CertPool
		if cluster.TLSConfig.CACertPath != "" {
			caCert, err := os.ReadFile(cluster.TLSConfig.CACertPath)
			if err != nil {
				panic(fmt.Sprintf("Unable to read CA cert file: %v", err))
			}

			caCertPool = x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
		}

		tlsConfig := &tls.Config{
			RootCAs:            caCertPool,
			InsecureSkipVerify: cluster.TLSConfig.SkipVerify,
		}

		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = tlsConfig
	}

	if cluster.SASLConfig.AuthMethod == config.AuthMethodSASLPlaintext {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		cfg.Net.SASL.User = cluster.SASLConfig.Username
		cfg.Net.SASL.Password = cluster.SASLConfig.Password
	}

	cfg.Net.DialTimeout = 5 * time.Second
	cfg.Net.ReadTimeout = 5 * time.Second
	cfg.Net.WriteTimeout = 5 * time.Second

	return cfg
}

func NewSaramaKadmin(cluster *config.Cluster) (Kadmin, error) {
	cfg := ToSaramaCfg(cluster)

	client, err := sarama.NewClient(cluster.BootstrapServers, cfg)
	if err != nil {
		return nil, err
	}

	admin, err := sarama.NewClusterAdmin(cluster.BootstrapServers, cfg)
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
		addrs:    cluster.BootstrapServers,
		producer: producer,
		config:   cfg,
	}, nil
}

func CheckKafkaConnectivity(cluster *config.Cluster) tea.Msg {
	connectedChan := make(chan bool)
	errChan := make(chan error)

	cfg := ToSaramaCfg(cluster)

	go doCheckConnectivity(cluster.BootstrapServers, cfg, errChan, connectedChan)

	return ConnCheckStartedMsg{
		Cluster:   cluster,
		Connected: connectedChan,
		Err:       errChan,
	}
}

func doCheckConnectivity(servers []string, config *sarama.Config, errChan chan error, connectedChan chan bool) {
	MaybeIntroduceLatency()
	c, err := sarama.NewClient(servers, config)
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
