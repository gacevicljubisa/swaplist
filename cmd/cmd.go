package cmd

import "github.com/spf13/cobra"

// nolint:gochecknoinits
func init() {
	cobra.EnableCommandSorting = false
}

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

func (c *command) Execute() (err error) {
	return c.root.Execute()
}

// Execute parses command line arguments and runs appropriate functions.
func Execute() (err error) {
	c, err := newCommand()
	if err != nil {
		return err
	}
	return c.Execute()
}
