package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_root(t *testing.T) {
	t.Run("root command", func(t *testing.T) {
		testRootCmd := NewRootCommand()
		assert.Equal(t, "upswake", testRootCmd.Use, "root command should be 'upswake'")
		assert.Equal(t, "UPSWake sends Wake on LAN packets based on a UPS's status", testRootCmd.Short, "root command short description mismatch")
		assert.Contains(t, testRootCmd.Long, "UPSWake sends Wake on LAN packets to target servers", "root command long description mismatch")
	})
}
