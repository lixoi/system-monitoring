package daemon

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "sysstatssvc",
	Short: "System statistic",
}

func init() {
	RootCmd.AddCommand(GrpcServerCmd)
}
