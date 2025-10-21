package main

import (
	"fmt"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/ups"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type jsonCMD struct {
	logger *zap.SugaredLogger
}

func NewJSONCommand(logger *zap.SugaredLogger) *cobra.Command {
	jc := &jsonCMD{logger: logger}
	cmd := &cobra.Command{
		Use:   "json",
		Short: "Retrieve JSON from a NUT server",
		Long: `Retrieve JSON from a NUT server and print it to stdout

This is useful for testing the connection to a NUT server
and for creating rego rules for waking a target`,
		Example: `  upswake json --host 192.168.1.66 --port 3493
  upswake json -H ups.example.com -P 3493 -u myuser -p mypass`,
		RunE: jc.JSONRunE,
	}
	setupJSONFlags(cmd)
	return cmd
}

func (j *jsonCMD) JSONRunE(cmd *cobra.Command, _ []string) error {
	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		j.logger.Errorf("could not get port: %s", err)
		return err
	}
	nutServer := entity.NutServer{
		Name:     "test",
		Host:     cmd.Flag("host").Value.String(),
		Port:     port,
		Username: cmd.Flag("username").Value.String(),
		Password: cmd.Flag("password").Value.String(),
	}

	upsData, err := ups.GetJSON(&nutServer)
	if err != nil {
		j.logger.Errorf("failed to get JSON: %s", err)
		return err
	}
	fmt.Println(upsData)
	return nil
}

func setupJSONFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("username", "u", "anonymous", "Username for the NUT server")
	cmd.Flags().StringP("password", "p", "anonymous", "Password for the NUT server")
	cmd.Flags().StringP("host", "H", "", "Host address of the NUT server")
	cmd.Flags().IntP("port", "P", entity.DefaultNUTServerPort, "Port number of the NUT server")
	_ = cmd.MarkFlagRequired("host")
}
