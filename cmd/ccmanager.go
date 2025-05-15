package main

// CCmanager - Cloud Control instance manager

import (
	"ccmanager/internal/adapters"
	"ccmanager/internal/models"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

func main() {
	var args struct {
		BasePath           []string `arg:"required,env:CCMANAGER_BASEPATH,separate" help:"Paths where to find CloudControl docker compose folders"`
		ContainerSeparator string   `default:"-" arg:"env:CCMANAGER_SEP" help:"Separator used in docker compose container names"`
	}
	arg.MustParse(&args)

	var items []list.Item

	d := adapters.DockerAdapter{}

	program := tea.NewProgram(models.NewMainModel(&d, args.BasePath, items))
	if _, err := program.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
