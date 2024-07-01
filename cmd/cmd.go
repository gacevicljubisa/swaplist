package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

type command struct {
	root *cobra.Command
}

func newCommand() (c *command, err error) {
	c = &command{
		root: &cobra.Command{
			Use:           "swaplist",
			Short:         "A tool to track transactions of a swarm node",
			Long:          "A tool to track transactions of a swarm node",
			SilenceErrors: true,
			SilenceUsage:  true,
		},
	}

	if err := c.initLimitCmd(); err != nil {
		return nil, err
	}

	if err := c.initFullCmd(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *command) Execute(ctx context.Context) (err error) {
	return c.root.ExecuteContext(ctx)
}

// Execute parses command line arguments and runs appropriate functions.
func Execute(ctx context.Context) (err error) {
	c, err := newCommand()
	if err != nil {
		return err
	}
	return c.Execute(ctx)
}
