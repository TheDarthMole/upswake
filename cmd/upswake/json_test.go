package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func Test_NewJSONCommand(t *testing.T) {
	t.Run("json command", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		sugar := logger.Sugar()
		jsonCmd := NewJSONCommand(sugar)
		assert.Equal(t, "json", jsonCmd.Use, "json command should be 'json'")
		assert.NotEmpty(t, jsonCmd.Short)
		assert.NotEmpty(t, jsonCmd.Long)
		assert.NotEmpty(t, jsonCmd.Example)
		assert.Equal(t, "anonymous", jsonCmd.Flags().Lookup("username").DefValue, "default username should be 'anonymous'")
		assert.Equal(t, "anonymous", jsonCmd.Flags().Lookup("password").DefValue, "default password should be 'anonymous'")
		assert.Empty(t, jsonCmd.Flags().Lookup("host").DefValue, "default host should be empty")
		assert.Equal(t, "3493", jsonCmd.Flags().Lookup("port").DefValue, "default port should be '3493'")
		assert.NotNil(t, jsonCmd.RunE, "json command RunE function should not be nil")
	})
}

func Test_JSONRunE(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		err  string
		out  string
	}{
		{
			name: "invalid port",
			in:   []string{"json", "--host", "localhost", "--username", "testuser", "--password", "testpass", "--port", "invalid"},
			out:  `invalid argument "invalid" for "-P, --port" flag`,
		},
		{
			name: "valid input but server not reachable",
			in:   []string{"json", "--host", "127.0.0.1", "--username", "testuser", "--password", "testpass", "--port", "1234"},
			err:  "could not connect to NUT server: dial tcp 127.0.0.1:1234",
		},
		{
			name: "valid cli args",
			in:   []string{"json", "--host", "localhost", "--username", "testuser", "--password", "testpass", "--port", "3493"},
		},
		{
			name: "missing host",
			in:   []string{},
			err:  `required flag(s) "host" not set`,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			output, err := executeCommandWithContext(t, NewJSONCommand, 1*time.Second, testCase.in...)

			if testCase.err != "" {
				assert.ErrorContains(t, err, testCase.err)
			}

			assert.Contains(t, output, testCase.out, "expected output not found")
		})
	}
}
