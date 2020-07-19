package cmd

import (
	"github.com/spf13/cobra"
)

func startCli(cmd *cobra.Command, opts OptionsCli) {
	if cap(opts.args) == 0 {
		cmd.Help()
	}
	// fmt.Println("args", args, endpoint)
}
