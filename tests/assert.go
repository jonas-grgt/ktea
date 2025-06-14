package tests

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TrimAndEqual(t *testing.T, actual string, expected string, stuff ...interface{}) {
	assert.Equal(t,
		strings.TrimSpace(expected),
		strings.TrimSpace(actual),
		stuff...,
	)
}
