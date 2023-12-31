package main

import (
	"fmt"

	"github.com/lixoi/system_stats_daemon/cmd/daemon"
)

func main() {
	if err := daemon.RootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
	}
}
