package internal

import "github.com/charmbracelet/lipgloss"

var (
	AppStyle = lipgloss.NewStyle().Padding(1, 2)

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	SelectedItemTitleStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
				Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
				Padding(0, 0, 0, 1)

	SelectedItemDescriptionStyle = SelectedItemTitleStyle.Copy().
					Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})

	ErrorMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#B5041a", Dark: "#B5041a"}).
				Render

	InfoBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(0, 1)

	StatusLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#aa0000")).
			Foreground(lipgloss.Color("#ffffff")).
			PaddingLeft(1)

	Labels = map[string]string{
		"AWS": lipgloss.NewStyle().
			Background(lipgloss.Color("#ff9900")).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 1).
			Render("AWS"),
		"Azure": lipgloss.NewStyle().
			Background(lipgloss.Color("#33b2e7")).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 1).
			Render("Azure"),
		"GCP": lipgloss.NewStyle().
			Background(lipgloss.Color("#f04943")).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 1).
			Render("GCP"),
		"Tanzu": lipgloss.NewStyle().
			Background(lipgloss.Color("#82c13d")).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 1).
			Render("Tanzu"),
		"Simple": lipgloss.NewStyle().
			Background(lipgloss.Color("#aaaaaa")).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 1).
			Render("Simple"),
		"Errpr": lipgloss.NewStyle().
			Background(lipgloss.Color("#ff0000")).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 1).
			Render("Error"),
	}
)
