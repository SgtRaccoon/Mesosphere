package main

import (
	"fmt"
	"os"

	"github.com/mesosphere/mesosphere/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "mesosphere: %v\n", err)
		os.Exit(1)
	}
}
