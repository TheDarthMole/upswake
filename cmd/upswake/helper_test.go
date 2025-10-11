package main

import (
	"bytes"
	"context"

	"github.com/spf13/cobra"
)

func executeCommandWithContextC(ctx context.Context, root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()

	return c, buf.String(), err
}
