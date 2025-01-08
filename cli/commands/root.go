package commands

import (
	"fmt"
	"os"

	"github.com/brainmaniac/gogogo/cli/generator"
)

func Execute() error {
	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	command := os.Args[1]

	switch command {
	case "new":
		if len(os.Args) < 3 {
			return fmt.Errorf("project name required")
		}
		return generator.GenerateProject(os.Args[2])
	default:
		printUsage()
	}

	return nil
}

func printUsage() {
	fmt.Println("GoGoGo - An opinionated Go framework")
	fmt.Println("\nUsage:")
	fmt.Println("  gogogo new [project-name]    Create a new project")
}
