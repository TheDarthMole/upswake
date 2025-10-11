package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_root(t *testing.T) {
	t.Run("root command", func(t *testing.T) {
		assert.Equal(t, "upswake", rootCmd.Use, "root command should be 'upswake'")
		assert.Equal(t, "UPSWake sends Wake on LAN packets based on a UPS's status", rootCmd.Short, "root command short description mismatch")
		assert.Contains(t, rootCmd.Long, "UPSWake sends Wake on LAN packets to target servers", "root command long description mismatch")
	})
}

//func Test_main(t *testing.T) {
//	t.Run("main function", func(t *testing.T) {
//		// Since main() calls rootCmd.Execute(), we can test if the rootCmd is set up correctly
//		main()
//		assert.NotNil(t, rootCmd, "rootCmd should not be nil")
//		assert.NotNil(t, sugar, "sugar logger should not be nil")
//	})
//}
