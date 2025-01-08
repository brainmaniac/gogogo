package main

import (
	"log"
	"os"

	"github.com/brainmaniac/gogogo/cli/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
