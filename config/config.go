package config

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"os"
)

type AuthMethod int

type SecurityProtocol string

const (
	NoneAuthMethod            AuthMethod       = 0
	SASLAuthMethod            AuthMethod       = 1
	SSLSecurityProtocol       SecurityProtocol = "SSL"
	PlaintextSecurityProtocol SecurityProtocol = "PLAIN_TEXT"
)

type SASLConfig struct {
	Username         string           `yaml:"username"`
	Password         string           `yaml:"password"`
	SecurityProtocol SecurityProtocol `yaml:"securityProtocol"`
}

type SchemaRegistryConfig struct {
	Url      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Cluster struct {
	Name             string                `yaml:"name"`
	Color            string                `yaml:"color"`
	Active           bool                  `yaml:"active"`
	BootstrapServers []string              `yaml:"servers"`
	SASLConfig       *SASLConfig           `yaml:"sasl"`
	SchemaRegistry   *SchemaRegistryConfig `yaml:"schema-registry"`
}

func (c *Cluster) HasSchemaRegistry() bool {
	return c.SchemaRegistry != nil
}

type Config struct {
	Clusters []Cluster `yaml:"clusters"`
	ConfigIO ConfigIO  `yaml:"-"`
}

func (c *Config) HasEnvs() bool {
	return len(c.Clusters) > 0
}

type SchemaRegistryDetails struct {
	Url      string
	Username string
	Password string
}

type RegistrationDetails struct {
	Name             string
	Color            string
	Host             string
	AuthMethod       AuthMethod
	SecurityProtocol SecurityProtocol
	NewName          *string
	Username         string
	Password         string
	SchemaRegistry   *SchemaRegistryDetails
}

type ClusterDeletedMsg struct {
	Name string
}

type ClusterRegisteredMsg struct {
	Cluster *Cluster
}

type ClusterRegisterer interface {
	RegisterCluster(d RegistrationDetails) tea.Msg
}

// RegisterCluster registers a new cluster or updates an existing one in the Config.
//
// If a cluster with the same name exists, it updates its details while retaining the "Active" status (the active param
// in that case is ignored) and optionally renaming it. Otherwise, it adds the cluster to the Config.
//
// It returns a ClusterRegisteredMsg with the registered cluster.
func (c *Config) RegisterCluster(details RegistrationDetails) tea.Msg {
	cluster := Cluster{
		Name:             details.Name,
		Color:            details.Color,
		BootstrapServers: []string{details.Host},
	}

	// When no clusters exist yet, the first one created becomes the active one by default.
	if len(c.Clusters) == 0 {
		cluster.Active = true
	}

	if details.AuthMethod == SASLAuthMethod {
		cluster.SASLConfig = &SASLConfig{
			Username:         details.Username,
			Password:         details.Password,
			SecurityProtocol: details.SecurityProtocol,
		}
	}

	if details.SchemaRegistry != nil {
		cluster.SchemaRegistry = &SchemaRegistryConfig{
			Url:      details.SchemaRegistry.Url,
			Username: details.SchemaRegistry.Username,
			Password: details.SchemaRegistry.Password,
		}
	}

	// did the newly registered cluster update an existing one
	var isUpdated bool

	for i := range c.Clusters {
		if c.Clusters[i].Name == details.Name {
			isActive := c.Clusters[i].Active
			cluster.Active = isActive
			c.Clusters[i] = cluster
			if details.NewName != nil {
				c.Clusters[i].Name = *details.NewName
			}
			isUpdated = true
			break
		}
	}

	// not updated, add new cluster
	if !isUpdated {
		c.Clusters = append(c.Clusters, cluster)
	}

	c.flush()

	return ClusterRegisteredMsg{&cluster}
}

func (c *Config) ActiveCluster() *Cluster {
	for _, c := range c.Clusters {
		if c.Active {
			return &c
		}
	}
	if len(c.Clusters) > 0 {
		return &c.Clusters[0]
	}
	return nil
}

func (c *Config) flush() {
	if err := c.ConfigIO.write(c); err != nil {
		fmt.Println("Unable to write config file")
		os.Exit(-1)
	}
}

func (c *Config) SwitchCluster(name string) *Cluster {
	var activeCluster *Cluster

	for i := range c.Clusters {
		// deactivate all clusters
		c.Clusters[i].Active = false
		if c.Clusters[i].Name == name {
			c.Clusters[i].Active = true
			activeCluster = &c.Clusters[i]
			log.Debug("switched to cluster " + name)
		}
	}

	c.flush()

	return activeCluster
}

func (c *Config) DeleteCluster(name string) {
	// Find the index of the cluster to delete
	index := -1
	for i, cluster := range c.Clusters {
		if cluster.Name == name {
			index = i
			break
		}
	}

	if index == -1 {
		log.Warn("cluster not found: " + name)
		return
	}

	// Check if the cluster to delete is active
	isActive := c.Clusters[index].Active

	// Remove the cluster
	c.Clusters = append(c.Clusters[:index], c.Clusters[index+1:]...)

	// Reactivate the first cluster if needed
	if isActive && len(c.Clusters) > 0 {
		c.Clusters[0].Active = true
	}

	c.flush()

	log.Debug("deleted cluster: " + name)
}

func (c *Config) FindClusterByName(name string) *Cluster {
	for _, cluster := range c.Clusters {
		if cluster.Name == name {
			return &cluster
		}
	}
	return nil
}

type LoadedMsg struct {
	Config *Config
}

func New(configIO ConfigIO) *Config {
	config, err := configIO.read()
	if err != nil {
		fmt.Println("Error reading config file:", err)
		os.Exit(-1)
	}
	config.ConfigIO = configIO
	return config
}

func ReLoadConfig() tea.Msg {
	return nil
}