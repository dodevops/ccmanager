package models

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// ConfirmItem is an implementation of list.Item for confirmation boxes. It basically
// holds the title ("yes"/"no") and the tea.Cmd to run if the option was chosen
type ConfirmItem struct {
	title   string
	command tea.Cmd
}

var _ list.Item = ConfirmItem{}

func (c ConfirmItem) FilterValue() string {
	return c.title
}

func (c ConfirmItem) Title() string {
	return c.title
}

func (c ConfirmItem) Description() string {
	return ""
}
