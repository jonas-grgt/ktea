package main

import (
	"errors"
	"ktea/config"
	"ktea/kadmin"
	"testing"

	"github.com/stretchr/testify/assert"
)

type closeRecordingKadmin struct {
	kadmin.MockKadmin
	closed bool
}

func (k *closeRecordingKadmin) Close() error {
	k.closed = true
	return nil
}

func TestRecreateAdminClients(t *testing.T) {
	var created []*closeRecordingKadmin
	var failNext bool
	instantiator := func(*config.Cluster) (kadmin.Kadmin, error) {
		if failNext {
			return nil, errors.New("unreachable")
		}
		ka := &closeRecordingKadmin{}
		created = append(created, ka)
		return ka, nil
	}
	m := &Model{kaInstantiator: instantiator}
	cluster := &config.Cluster{Name: "c"}

	assert.NoError(t, m.recreateAdminClients(cluster))
	assert.NoError(t, m.recreateAdminClients(cluster))
	assert.True(t, created[0].closed, "previous cluster connection is closed")
	assert.False(t, created[1].closed)

	failNext = true
	assert.Error(t, m.recreateAdminClients(cluster))
	assert.False(t, created[1].closed, "a failed switch keeps the working connection")
	assert.Same(t, created[1], m.ka)
}
