package kadmin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetClusterConfig(t *testing.T) {
	t.Run("Get cluster config", func(t *testing.T) {
		// when
		msg := ka.GetClusterConfig()
		startedMsg, ok := msg.(ClusterConfigStartedMsg)
		assert.True(t, ok, "expected ClusterConfigStartedMsg")

		// then
		select {
		case result := <-startedMsg.Configs:
			assert.Len(t, result.Brokers, 1, "Expected 1 broker")
			assert.NotNil(t, result.Brokers[0].ID, "Broker ID should not be nil")
			assert.NotEmpty(t, result.Brokers[0].Address, "Broker address should not be empty")
		case err := <-startedMsg.Err:
			t.Fatal("Error while getting cluster config", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out waiting for cluster config")
		}
	})
}

func TestGetBrokerConfig(t *testing.T) {
	t.Run("Get broker config", func(t *testing.T) {
		// given
		// Get the broker ID from the cluster config
		msg := ka.GetClusterConfig()
		startedMsg, ok := msg.(ClusterConfigStartedMsg)
		assert.True(t, ok, "expected ClusterConfigStartedMsg")

		var brokerID int32
		select {
		case result := <-startedMsg.Configs:
			assert.Len(t, result.Brokers, 1, "Expected 1 broker")
			brokerID = result.Brokers[0].ID
		case err := <-startedMsg.Err:
			t.Fatal("Error while getting cluster config for broker test", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out waiting for cluster config for broker test")
		}

		// when
		brokerMsg := ka.GetBrokerConfig(brokerID)
		brokerStartedMsg, ok := brokerMsg.(BrokerConfigListingStartedMsg)
		assert.True(t, ok, "expected BrokerConfigListingStartedMsg")

		// then
		select {
		case result := <-brokerStartedMsg.Configs:
			assert.Equal(t, brokerID, result.ID)
			assert.NotEmpty(t, result.Configs, "Expected broker config to not be empty")
		case err := <-brokerStartedMsg.Err:
			t.Fatal("Error while getting broker config", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out waiting for broker config")
		}
	})
}

func TestGetBrokerConfig_Error(t *testing.T) {
	t.Run("Get broker config with non-existent broker ID", func(t *testing.T) {
		// when
		brokerMsg := ka.GetBrokerConfig(999)
		brokerStartedMsg, ok := brokerMsg.(BrokerConfigListingStartedMsg)
		assert.True(t, ok, "expected BrokerConfigListingStartedMsg")

		// then
		select {
		case <-brokerStartedMsg.Configs:
			t.Fatal("Expected an error but got broker config instead")
		case err := <-brokerStartedMsg.Err:
			assert.Error(t, err, "Expected an error")
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out waiting for broker config error")
		}
	})
}
