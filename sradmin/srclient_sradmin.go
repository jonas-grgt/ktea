package sradmin

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/riferrei/srclient"
	"ktea/config"
	"net/http"
	"os"
	"sync"
)

type DefaultSrClient struct {
	client      *srclient.SchemaRegistryClient
	subjects    []Subject
	mu          sync.RWMutex
	schemaCache map[int]Schema
}

type Client interface {
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

func (s *DefaultSrClient) CreateSchema(details SubjectCreationDetails) tea.Msg {
	createdChan := make(chan bool)
	errChan := make(chan error)

	go s.doCreateSchema(details, createdChan, errChan)

	return SchemaCreationStartedMsg{
		createdChan,
		errChan,
	}
}

func (s *DefaultSrClient) doCreateSchema(details SubjectCreationDetails, createdChan chan bool, errChan chan error) {
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

	tlsConfig := &tls.Config{
		InsecureSkipVerify: registry.TLSConfig.SkipVerify,
	}

	if registry.TLSConfig.CACertPath != "" {
		caCert, err := os.ReadFile(registry.TLSConfig.CACertPath)
		if err != nil {
			panic(fmt.Sprintf("Unable to read CA cert file: %v", err))
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			panic("Failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	if registry.TLSConfig.ClientCert != "" && registry.TLSConfig.ClientKey != "" {
		cert, err := tls.LoadX509KeyPair(registry.TLSConfig.ClientCert, registry.TLSConfig.ClientKey)
		if err != nil {
			panic(fmt.Sprintf("Unable to load client certificate: %v", err))
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	transport := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsConfig,
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

func New(registryConfig *config.SchemaRegistryConfig) *DefaultSrClient {
	client := createHttpClient(registryConfig)
	return &DefaultSrClient{
		client: srclient.NewSchemaRegistryClient(registryConfig.Url, srclient.WithClient(client)),
	}
}
