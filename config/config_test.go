package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	userJohn = "john"
	pwdJohn  = "doe"

	userJane = "jane"
	pwdJane  = "do3"
)

func TestConfig(t *testing.T) {

	t.Run("Registering a SASL cluster", func(t *testing.T) {
		// given
		config := New(&InMemoryConfigIO{})

		// when
		config.RegisterCluster(RegistrationDetails{
			Name:       "prd",
			Color:      "#880808",
			Host:       "localhost:9092",
			AuthMethod: AuthMethodSASLPlaintext,
			Username:   userJohn,
			Password:   "test123",
		})

		// then
		assert.Equal(t, config.Clusters[0].Color, "#880808")
		assert.Equal(t, config.Clusters[0].BootstrapServers, []string{"localhost:9092"})
		assert.Equal(t, config.Clusters[0].SASLConfig.AuthMethod, AuthMethodSASLPlaintext)
		assert.Equal(t, config.Clusters[0].SASLConfig.Username, userJohn)
		assert.Equal(t, config.Clusters[0].SASLConfig.Password, "test123")
		assert.Equal(t, config.Clusters[0].TLSConfig, TLSConfig{
			Enable:     false,
			SkipVerify: false,
			CACertPath: "",
		})
	})

	t.Run("Registering a SASL cluster with SSL", func(t *testing.T) {
		// given
		config := New(&InMemoryConfigIO{})

		// when
		config.RegisterCluster(RegistrationDetails{
			Name:       "prd",
			Color:      "#880808",
			Host:       "localhost:9092",
			AuthMethod: AuthMethodSASLPlaintext,
			Username:   userJohn,
			Password:   "test123",
			TLSConfig: TLSConfig{
				Enable:     true,
				SkipVerify: false,
				CACertPath: "~/.certs/ca.crt",
			},
		})

		// then
		assert.Equal(t, config.Clusters[0].Color, "#880808")
		assert.Equal(t, config.Clusters[0].BootstrapServers, []string{"localhost:9092"})
		assert.Equal(t, config.Clusters[0].SASLConfig.AuthMethod, AuthMethodSASLPlaintext)
		assert.Equal(t, config.Clusters[0].SASLConfig.Username, userJohn)
		assert.Equal(t, config.Clusters[0].SASLConfig.Password, "test123")
		assert.Equal(t, config.Clusters[0].TLSConfig, TLSConfig{
			Enable:     true,
			SkipVerify: false,
			CACertPath: "~/.certs/ca.crt",
		})
	})

	t.Run("Registering a SASL_SSL Cluster with Schema Registry ", func(t *testing.T) {
		// given
		config := New(&InMemoryConfigIO{})

		// when
		config.RegisterCluster(RegistrationDetails{
			Name:       "prd",
			Color:      "#880808",
			Host:       "localhost:9092",
			AuthMethod: AuthMethodSASLPlaintext,
			Username:   userJohn,
			Password:   "test123",
			SchemaRegistry: &SchemaRegistryDetails{
				Url:      "https://sr:1923",
				Username: "srJohn",
				Password: "srTest123",
			},
		})

		// then
		assert.Equal(t, config.Clusters[0].Color, "#880808")
		assert.Equal(t, config.Clusters[0].BootstrapServers, []string{"localhost:9092"})
		assert.Equal(t, config.Clusters[0].SASLConfig.AuthMethod, AuthMethodSASLPlaintext)
		assert.Equal(t, config.Clusters[0].SASLConfig.Username, userJohn)
		assert.Equal(t, config.Clusters[0].SASLConfig.Password, "test123")
		assert.Equal(t, config.Clusters[0].SchemaRegistry.Url, "https://sr:1923")
		assert.Equal(t, config.Clusters[0].SchemaRegistry.Username, "srJohn")
		assert.Equal(t, config.Clusters[0].SchemaRegistry.Password, "srTest123")
	})

	t.Run("Registering a first cluster activates it by default", func(t *testing.T) {
		// given
		config := New(&InMemoryConfigIO{})

		// when
		config.RegisterCluster(RegistrationDetails{
			Name:       "prd",
			Color:      "#880801",
			Host:       "localhost:9093",
			AuthMethod: AuthMethodNone,
		})

		// then
		assert.Equal(t, config.Clusters[0].Name, "prd")
		assert.Equal(t, config.Clusters[0].Active, true)
	})

	t.Run("Registering an existing cluster updates it", func(t *testing.T) {
		// given
		config := New(&InMemoryConfigIO{})
		config.RegisterCluster(RegistrationDetails{
			Name:       "prd",
			Color:      "#880808",
			Host:       "localhost:9092",
			AuthMethod: AuthMethodNone,
		})

		// when
		config.RegisterCluster(RegistrationDetails{
			Name:       "prd",
			Color:      "#880801",
			Host:       "localhost:9093",
			AuthMethod: AuthMethodNone,
		})

		// then
		assert.Equal(t, config.Clusters[0].Color, "#880801")
		assert.Equal(t, config.Clusters[0].BootstrapServers, []string{"localhost:9093"})
	})

	t.Run("Registering an existing active cluster keeps it active", func(t *testing.T) {
		// given
		config := New(&InMemoryConfigIO{})
		config.RegisterCluster(RegistrationDetails{
			Name:       "prd",
			Color:      "#880808",
			Host:       "localhost:9092",
			AuthMethod: AuthMethodNone,
		})

		// when
		newName := "PRD"
		clusterRegisteredMsg := config.RegisterCluster(RegistrationDetails{
			NewName:    &newName,
			Name:       "prd",
			Color:      "#880801",
			Host:       "localhost:9093",
			AuthMethod: AuthMethodNone,
		}).(ClusterRegisteredMsg)

		// then
		assert.Equal(t, config.Clusters[0].Name, "PRD")
		assert.Equal(t, clusterRegisteredMsg.Cluster.Active, true)
		assert.Equal(t, config.Clusters[0].Name, "PRD")
		assert.Equal(t, clusterRegisteredMsg.Cluster.Active, true)
	})

	t.Run("Registering an existing cluster with new Name", func(t *testing.T) {
		// given
		config := New(&InMemoryConfigIO{})
		config.RegisterCluster(RegistrationDetails{
			Name:       "prd",
			Color:      "#880808",
			Host:       "localhost:9093",
			AuthMethod: AuthMethodNone,
		})

		// when
		updatedName := "prod"
		config.RegisterCluster(RegistrationDetails{
			Name:       "prd",
			NewName:    &updatedName,
			Color:      "#880808",
			Host:       "localhost:9093",
			AuthMethod: AuthMethodNone,
		})

		// then
		assert.Len(t, config.Clusters, 1)
		assert.Equal(t, config.Clusters[0].Name, "prod")
		assert.Equal(t, config.Clusters[0].BootstrapServers, []string{"localhost:9093"})
	})

	t.Run("Registering an existing cluster with first Kafka Connect Cluster", func(t *testing.T) {
		// given
		config := New(&InMemoryConfigIO{})
		config.RegisterCluster(RegistrationDetails{
			Name:       "prd",
			Color:      "#880808",
			Host:       "localhost:9092",
			AuthMethod: AuthMethodNone,
		})

		// when
		config.RegisterCluster(RegistrationDetails{
			Name:       "prd",
			Color:      "#880808",
			Host:       "localhost:9092",
			AuthMethod: AuthMethodNone,
			KafkaConnectClusters: []KafkaConnectClusterDetails{
				{
					Name:     "s3-sink",
					Url:      "http://localhost:8083",
					Username: &userJohn,
					Password: &pwdJohn,
				},
			},
		})

		// then
		assert.Equal(t, config.Clusters[0].Color, "#880808")
		assert.Equal(t, config.Clusters[0].BootstrapServers, []string{"localhost:9092"})
		// and
		assert.Len(t, config.Clusters[0].KafkaConnectClusters, 1)
		assert.Equal(t, config.Clusters[0].KafkaConnectClusters[0], KafkaConnectConfig{
			Name:     "s3-sink",
			Url:      "http://localhost:8083",
			Username: &userJohn,
			Password: &pwdJohn,
		})
	})

	t.Run("Registering an existing cluster with second Kafka Connect Cluster", func(t *testing.T) {
		// given
		config := New(&InMemoryConfigIO{})
		config.RegisterCluster(RegistrationDetails{
			Name:       "prd",
			Color:      "#880808",
			Host:       "localhost:9092",
			AuthMethod: AuthMethodNone,
			KafkaConnectClusters: []KafkaConnectClusterDetails{
				{
					Name:     "s3-sink",
					Url:      "http://localhost:8083",
					Username: &userJohn,
					Password: &pwdJohn,
				},
			},
		})

		// when
		config.RegisterCluster(RegistrationDetails{
			Name:       "prd",
			Color:      "#880808",
			Host:       "localhost:9092",
			AuthMethod: AuthMethodNone,
			KafkaConnectClusters: []KafkaConnectClusterDetails{
				{
					Name:     "s3-sink",
					Url:      "http://localhost:8083",
					Username: &userJohn,
					Password: &pwdJohn,
				},
				{
					Name:     "s4-sink",
					Url:      "http://localhost:8084",
					Username: &userJane,
					Password: &pwdJane,
				},
			},
		})

		// then
		assert.Equal(t, config.Clusters[0].Color, "#880808")
		assert.Equal(t, config.Clusters[0].BootstrapServers, []string{"localhost:9092"})
		// and
		assert.Len(t, config.Clusters[0].KafkaConnectClusters, 2)
		assert.Contains(t, config.Clusters[0].KafkaConnectClusters, KafkaConnectConfig{
			Name:     "s3-sink",
			Url:      "http://localhost:8083",
			Username: &userJohn,
			Password: &pwdJohn,
		})
		assert.Contains(t, config.Clusters[0].KafkaConnectClusters, KafkaConnectConfig{
			Name:     "s4-sink",
			Url:      "http://localhost:8084",
			Username: &userJane,
			Password: &pwdJane,
		})
	})

	t.Run("Get primary cluster", func(t *testing.T) {
		t.Run("When defined", func(t *testing.T) {
			config := New(&InMemoryConfigIO{})
			config.RegisterCluster(RegistrationDetails{
				Name:       "prd",
				Color:      "#880808",
				Host:       "localhost:9092",
				AuthMethod: AuthMethodNone,
			})

			cluster := config.ActiveCluster()

			assert.Equal(t, &Cluster{
				Name:             "prd",
				Color:            "#880808",
				Active:           true,
				BootstrapServers: []string{"localhost:9092"},
				SASLConfig: SASLConfig{
					Username:   "",
					Password:   "",
					AuthMethod: AuthMethodNone,
				},
				TLSConfig: TLSConfig{
					Enable:     false,
					SkipVerify: false,
					CACertPath: "",
				},
			}, cluster)

		})

		t.Run("When not defined take first", func(t *testing.T) {
			config := &Config{
				Clusters: []Cluster{
					{
						Name:             "prd1",
						Color:            "#880808",
						Active:           false,
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig:       SASLConfig{AuthMethod: AuthMethodNone},
					},
					{
						Name:             "prd2",
						Color:            "#880808",
						Active:           false,
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig:       SASLConfig{AuthMethod: AuthMethodNone},
					},
				},
			}

			cluster := config.ActiveCluster()

			assert.Equal(t, &Cluster{
				Name:             "prd1",
				Color:            "#880808",
				Active:           false,
				BootstrapServers: []string{"localhost:9092"},
				SASLConfig:       SASLConfig{AuthMethod: AuthMethodNone},
			}, cluster)

		})
	})

	t.Run("Delete existing cluster", func(t *testing.T) {
		// given
		config := New(&InMemoryConfigIO{})
		config.RegisterCluster(RegistrationDetails{
			Name:       "prd",
			Color:      "#880808",
			Host:       "localhost:9092",
			AuthMethod: AuthMethodNone,
		})
		config.RegisterCluster(RegistrationDetails{
			Name:       "tst",
			Color:      "#880808",
			Host:       "localhost:9093",
			AuthMethod: AuthMethodNone,
		})
		config.RegisterCluster(RegistrationDetails{
			Name:       "uat",
			Color:      "#880808",
			Host:       "localhost:9094",
			AuthMethod: AuthMethodNone,
		})

		// when
		config.DeleteCluster("tst")

		// then
		for _, cluster := range config.Clusters {
			if cluster.Name == "tst" {
				t.Fatal("tst found and not deleted")
			}
		}
	})

	t.Run("Delete existing active cluster", func(t *testing.T) {
		t.Run("activate first one", func(t *testing.T) {
			// given
			config := New(&InMemoryConfigIO{})
			config.RegisterCluster(RegistrationDetails{
				Name:       "prd",
				Color:      "#880808",
				Host:       "localhost:9092",
				AuthMethod: AuthMethodNone,
			})
			config.RegisterCluster(RegistrationDetails{
				Name:       "tst",
				Color:      "#880808",
				Host:       "localhost:9093",
				AuthMethod: AuthMethodNone,
			})

			// when
			config.DeleteCluster("tst")

			// then
			for _, cluster := range config.Clusters {
				if cluster.Name == "tst" {
					t.Fatal("tst found and not deleted")
				}
				if cluster.Name == "prd" && !cluster.Active {
					t.Fatal("prd found but not activate")
				}
			}
		})

		t.Run("when only cluster available", func(t *testing.T) {
			// given
			config := New(&InMemoryConfigIO{})
			config.RegisterCluster(RegistrationDetails{
				Name:       "prd",
				Color:      "#880808",
				Host:       "localhost:9092",
				AuthMethod: AuthMethodNone,
			})

			// when
			config.DeleteCluster("prd")

			// then
			for _, cluster := range config.Clusters {
				if cluster.Name == "prd" {
					t.Fatal("prd found and not deleted")
				}
			}
		})
	})

	t.Run("Find existing cluster by Name", func(t *testing.T) {
		// given
		config := New(&InMemoryConfigIO{})
		config.RegisterCluster(RegistrationDetails{
			Name:       "prd",
			Color:      "#880808",
			Host:       "localhost:9092",
			AuthMethod: AuthMethodNone,
		})
		config.RegisterCluster(RegistrationDetails{
			Name:       "tst",
			Color:      "#880808",
			Host:       "localhost:9093",
			AuthMethod: AuthMethodNone,
		})
		config.RegisterCluster(RegistrationDetails{
			Name:       "uat",
			Color:      "#880808",
			Host:       "localhost:9094",
			AuthMethod: AuthMethodNone,
		})

		// when
		cluster := config.FindClusterByName("tst")

		// then
		assert.Equal(t, cluster.Name, "tst")
		assert.Equal(t, cluster.BootstrapServers, []string{"localhost:9093"})
	})

	t.Run("Find non existing cluster by Name", func(t *testing.T) {
		// given
		config := New(&InMemoryConfigIO{})
		config.RegisterCluster(RegistrationDetails{
			Name:       "prd",
			Color:      "#880808",
			Host:       "localhost:9092",
			AuthMethod: AuthMethodNone,
		})
		config.RegisterCluster(RegistrationDetails{
			Name:       "tst",
			Color:      "#880808",
			Host:       "localhost:9093",
			AuthMethod: AuthMethodNone,
		})
		config.RegisterCluster(RegistrationDetails{
			Name:       "uat",
			Color:      "#880808",
			Host:       "localhost:9094",
			AuthMethod: AuthMethodNone,
		})

		// when
		cluster := config.FindClusterByName("tsty")

		// then
		assert.Nil(t, cluster)
	})
}
