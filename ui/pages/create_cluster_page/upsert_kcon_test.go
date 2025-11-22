package create_cluster_page

import (
	"ktea/config"
	"ktea/kcadmin"
	"ktea/tests"
	"ktea/ui/components/cmdbar"
	"ktea/ui/tabs"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestUpsertKcModel(t *testing.T) {

	ktx := tests.NewKontext()

	t.Run("Immediately show form when no clusters registered", func(t *testing.T) {
		m := NewUpsertKcModel(
			tabs.NewMockClustersTabNavigator(),
			ktx,
			nil,
			[]config.KafkaConnectConfig{},
			kcadmin.NewMockConnChecker(),
			cmdbar.NewNotifierCmdBar("test"),
			mockKafkaConnectRegisterer,
		)

		render := m.View(ktx, tests.Renderer)

		assert.Contains(t, render, "Kafka Connect Name")
		assert.Contains(t, render, "Kafka Connect URL")
		assert.Contains(t, render, "Kafka Connect Username")
		assert.Contains(t, render, "Kafka Connect Password")

		t.Run("Tests connection upon creation", func(t *testing.T) {
			tests.UpdateKeys(m, "dev sink cluster")
			cmd := m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())

			tests.UpdateKeys(m, "http://localhost:8083")
			cmd = m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())

			tests.UpdateKeys(m, "jane")
			cmd = m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())

			tests.UpdateKeys(m, "doe")
			cmd = m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())

			msgs := tests.Submit(m)

			username := "jane"
			password := "doe"
			assert.Len(t, msgs, 1)
			assert.IsType(t, kcadmin.MockConnectionCheckedMsg{}, msgs[0])
			assert.Equal(t, &config.KafkaConnectConfig{
				Name:     "dev sink cluster",
				Url:      "http://localhost:8083",
				Username: &username,
				Password: &password,
			}, msgs[0].(kcadmin.MockConnectionCheckedMsg).Config)
		})

		t.Run("Register Kafka Connect Cluster upon successful connection", func(t *testing.T) {
			cmd := m.Update(kcadmin.ConnCheckSucceededMsg{})

			msgs := tests.ExecuteBatchCmd(cmd)

			assert.Len(t, msgs, 1)
			assert.IsType(t, mockKafkaConnectRegistered{}, msgs[0])
		})
	})

	t.Run("Set username and password transportOption nil when left empty", func(t *testing.T) {
		m := NewUpsertKcModel(
			tabs.NewMockClustersTabNavigator(),
			ktx,
			nil,
			[]config.KafkaConnectConfig{},
			kcadmin.NewMockConnChecker(),
			cmdbar.NewNotifierCmdBar("test"),
			mockKafkaConnectRegisterer,
		)

		tests.UpdateKeys(m, "dev sink cluster")
		cmd := m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())

		tests.UpdateKeys(m, "http://localhost:8083")
		cmd = m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())

		cmd = m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())

		cmd = m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())

		msgs := tests.Submit(m)

		assert.Len(t, msgs, 1)
		assert.IsType(t, kcadmin.MockConnectionCheckedMsg{}, msgs[0])
		assert.Equal(t, &config.KafkaConnectConfig{
			Name:     "dev sink cluster",
			Url:      "http://localhost:8083",
			Username: nil,
			Password: nil,
		}, msgs[0].(kcadmin.MockConnectionCheckedMsg).Config)

		details := m.clusterDetails()
		assert.Nil(t, details[0].Username)
		assert.Nil(t, details[0].Password)
	})

	t.Run("List kafka connect clusters when at least one is already registered", func(t *testing.T) {
		username := "jane"
		password := "doe"
		m := NewUpsertKcModel(
			tabs.NewMockClustersTabNavigator(),
			ktx,
			nil,
			[]config.KafkaConnectConfig{
				{
					Name:     "s3-sink",
					Url:      "http://localhost:8083",
					Username: &username,
					Password: &password,
				},
			},
			kcadmin.NewMockConnChecker(),
			cmdbar.NewNotifierCmdBar("test"),
			mockKafkaConnectRegisterer,
		)

		render := m.View(ktx, tests.Renderer)

		assert.NotContains(t, render, "Kafka Connect URL")
		assert.NotContains(t, render, "Kafka Connect Username")
		assert.NotContains(t, render, "Kafka Connect Password")

		assert.Contains(t, render, "s3-sink")
	})
}
