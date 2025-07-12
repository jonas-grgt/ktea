package sradmin

import (
	"encoding/base64"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/riferrei/srclient"
	"ktea/config"
	"net/http"
	"sync"
)

type DefaultSrAdmin struct {
	client      *srclient.SchemaRegistryClient
	subjects    []Subject
	mu          sync.RWMutex
	schemaCache map[int]Schema
}

type SrAdmin interface {
	SubjectDeleter
	SubjectLister
	SchemaCreator
	VersionLister
	SchemaFetcher
	GlobalCompatibilityLister
	LatestSchemaBySubjectFetcher
	SchemaDeleter
}

type ConnCheckSucceededMsg struct{}

type ConnCheckErrMsg struct {
	Err error
}

// ConnChecker is a function that checks a Schema Registry connection and returns a tea.Msg.
type ConnChecker func(c *config.SchemaRegistryConfig) tea.Msg

func (s *DefaultSrAdmin) GetSubjects() []Subject {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy of the slice to prevent external modifications
	subjectsCopy := make([]Subject, len(s.subjects))
	copy(subjectsCopy, s.subjects)
	return subjectsCopy
}

type SchemaCreationStartedMsg struct {
	created chan bool
	err     chan error
}

type SchemaCreatedMsg struct{}

type SchemaCreationErrMsg struct {
	Err error
}

func (msg *SchemaCreationStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-msg.created:
		return SchemaCreatedMsg{}
	case err := <-msg.err:
		return SchemaCreationErrMsg{err}
	}
}

func (s *DefaultSrAdmin) CreateSchema(details SubjectCreationDetails) tea.Msg {
	createdChan := make(chan bool)
	errChan := make(chan error)

	go s.doCreateSchema(details, createdChan, errChan)

	return SchemaCreationStartedMsg{
		createdChan,
		errChan,
	}
}

func (s *DefaultSrAdmin) doCreateSchema(details SubjectCreationDetails, createdChan chan bool, errChan chan error) {
	maybeIntroduceLatency()
	_, err := s.client.CreateSchema(details.Subject, details.Schema, srclient.Avro)
	if err != nil {
		errChan <- err
		return
	}
	createdChan <- true
}

func createHttpClient(registry *config.SchemaRegistryConfig) *http.Client {
	auth := registry.Username + ":" + registry.Password
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	client := &http.Client{
		Transport: roundTripperWithAuth{
			baseTransport: transport,
			authHeader:    authHeader,
		},
	}
	return client
}

type roundTripperWithAuth struct {
	baseTransport http.RoundTripper
	authHeader    string
}

// RoundTrip adds the Authorization header to every request
func (r roundTripperWithAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", r.authHeader)
	return r.baseTransport.RoundTrip(req)
}

func New(registryConfig *config.SchemaRegistryConfig) *DefaultSrAdmin {
	client := createHttpClient(registryConfig)
	return &DefaultSrAdmin{
		client: srclient.NewSchemaRegistryClient(registryConfig.Url, srclient.WithClient(client)),
	}
}
