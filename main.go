package main

import (
	"fmt"
	"os"

	"github.com/xcoulon/kubectl-terminate/cmd"
)

func main() {
	if err := cmd.TerminateCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
