package main

// CCmanager - Cloud Control instance manager

import (
	"ccmanager/internal/adapters"
	"ccmanager/internal/models"
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"os"
	"strings"
)

func main() {
	parser := argparse.NewParser("ccmanager", "CloudControl instance manager")
	basePath := parser.StringList("b", "basepath", &argparse.Options{
		Help:     "Paths where to find CloudControl docker compose folders. Can be set using a comma separated list in CCMANAGER_BASEPATH",
		Required: false,
	})

	if err := parser.Parse(os.Args); err != nil {
		log.Fatal(parser.Usage(err))
	}

	if len(*basePath) == 0 {
		if e, found := os.LookupEnv("CCMANAGER_BASEPATH"); found {
			*basePath = strings.Split(e, ",")
		}
	}

	if len(*basePath) == 0 {
		log.Fatal(parser.Usage("No basepath given. Please use CCMANAGER_BASEPATH or the --basepath argument"))
	}

	var items []list.Item

	d := adapters.DockerAdapter{}

	program := tea.NewProgram(models.NewMainModel(d, *basePath, items))
	if _, err := program.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
