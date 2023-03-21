package models

import (
	"ccmanager/internal"
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

// View renders the screen dependending on the model
func (m MainModel) View() string {
	if m.loadedItems {
		if m.ShowInfo {
			content := internal.InfoBoxStyle.Render(fmt.Sprintf(
				"Path: %s\nState: %s\nImage: %s",
				m.InfoItem.Path,
				m.InfoItem.StateDescription(),
				m.InfoItem.State.Image,
			))
			infoBox := lipgloss.JoinVertical(.5,
				internal.TitleStyle.
					Width(lipgloss.Width(content)).
					Render(fmt.Sprintf("Instance %s", m.InfoItem.Name)),
				content,
			)
			return lipgloss.JoinVertical(
				0,
				lipgloss.NewStyle().
					Width(m.Width).
					Height(m.Height-1).
					Align(lipgloss.Center, lipgloss.Center).
					Render(infoBox),
				internal.StatusLineStyle.Width(m.Width).Render("Press any key to return"),
			)
		} else if m.ShowLog {
			content := m.LogViewer.View()
			return lipgloss.JoinVertical(
				0,
				internal.TitleStyle.
					Width(m.Width).
					Render(fmt.Sprintf("Log of instance %s", m.InfoItem.Name)),
				content,
				internal.StatusLineStyle.Width(m.Width).Render("Press q or escape to return"),
			)
		} else if m.RunningConfirm {
			m.Confirm.SetWidth(lipgloss.Width(m.ConfirmPrompt))
			m.Confirm.SetHeight(4)
			content := internal.InfoBoxStyle.Render(
				lipgloss.JoinVertical(
					0.5,
					m.ConfirmPrompt+"\n",
					lipgloss.NewStyle().AlignHorizontal(lipgloss.Right).Render(m.Confirm.View()),
				),
			)
			infoBox := lipgloss.JoinVertical(.5,
				internal.TitleStyle.
					Width(lipgloss.Width(content)).
					Padding(0, 2).
					Render("Please select"),
				content,
			)
			return lipgloss.JoinVertical(
				0,
				lipgloss.NewStyle().
					Width(m.Width).
					Height(m.Height-1).
					Align(lipgloss.Center, lipgloss.Center).
					Render(infoBox),
				internal.StatusLineStyle.Width(m.Width).Render("Please select an option"),
			)
		} else if !m.ListDisabled {
			return internal.AppStyle.Render(m.List.View())
		} else {
			return fmt.Sprintf("%s\n\n", internal.AppStyle.Render(internal.TitleStyle.Render(m.List.Title)))
		}
	} else {
		return fmt.Sprintf("%s Loading instances", m.spinner.View())
	}
}
