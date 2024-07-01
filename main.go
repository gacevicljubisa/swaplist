package main

import (
	"fmt"
	"os"

	"github.com/gacevicljubisa/swaplist/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
