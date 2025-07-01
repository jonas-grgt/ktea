package kadmin

import (
	"ktea/config"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	PlainText SASLProtocol = 0
)

const (
	TopicResourceType = 2
)

type Kadmin interface {
	TopicCreator
	TopicDeleter
	TopicLister
	Publisher
	RecordReader
	OffsetLister
	CGroupLister
	CGroupDeleter
	ConfigUpdater
	TopicConfigLister
	SraSetter
	ClusterConfigLister
}

type ClusterConfigLister interface {
	GetClusterConfig() (ClusterConfig, error)
}

type BrokerConfig struct {
	ID      int32
	Addr    string
	Configs map[string]string
}

type ClusterConfig struct {
	Brokers []BrokerConfig
}

type ClusterConfigListingStartedMsg struct {
	Err     chan error
	Configs chan ClusterConfig
}

func (m *ClusterConfigListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case e := <-m.Err:
		return ClusterConfigListingErrorMsg{e}
	case c := <-m.Configs:
		return ClusterConfigListedMsg{c}
	}
}

type ClusterConfigListedMsg struct {
	Config ClusterConfig
}

type ClusterConfigListingErrorMsg struct {
	Err error
}

type ConnectionDetails struct {
	BootstrapServers []string
	SASLConfig       *SASLConfig
	SSLEnabled       bool
}

type SASLProtocol int

type SASLConfig struct {
	Username string
	Password string
	Protocol SASLProtocol
}

type GroupMember struct {
	MemberId   string
	ClientId   string
	ClientHost string
}

type KAdminErrorMsg struct {
	Error error
}

type ConnErrMsg struct {
	Err error
}

type Instantiator func(cd ConnectionDetails) (Kadmin, error)

type ConnChecker func(cluster *config.Cluster) tea.Msg

func SaramaInstantiator() Instantiator {
	return func(cd ConnectionDetails) (Kadmin, error) {
		return NewSaramaKadmin(cd)
	}
}
