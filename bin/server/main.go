package main

import (
	"os"

	"github.com/youzi-1122/bytebase/bin/server/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
