package models

import (
	"ccmanager/internal"
	"ccmanager/internal/adapters"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"time"
)

// ApplicationKeyMap holds a struct of Key.Binding entries that are used to control CCmanager
type ApplicationKeyMap struct {
	// ToggleTitleBar toggles the list title bar
	ToggleTitleBar key.Binding
	// ToggleStatusBar toggles the status bar
	ToggleStatusBar key.Binding
	// TogglePagination toggles the pagination bar
	TogglePagination key.Binding
	// ToggleHelpMenu toggles the long help menu
	ToggleHelpMenu key.Binding
	// Refresh refreshes the list of instances
	Refresh key.Binding
	// Run starts CloudControl in an instance
	Run key.Binding
	// Restart restarts CloudControl
	Restart key.Binding
	// OpenCCC calls the browser to open CCC
	OpenCCC key.Binding
	// Stop stops an instance
	Stop key.Binding
	// ShowLog shows the log of an instance
	ShowLog key.Binding
	// Info shows the information about an instance
	Info key.Binding
	// Start starts an instance
	Start key.Binding
}

func NewApplicationKeyMap() *ApplicationKeyMap {
	return &ApplicationKeyMap{
		ToggleTitleBar: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "toggle title"),
		),
		ToggleStatusBar: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "toggle status"),
		),
		TogglePagination: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle pagination"),
		),
		ToggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "refresh instances"),
		),
		Run: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "enter"),
		),
		Restart: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restart"),
		),
		OpenCCC: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "open CCC"),
		),
		Stop: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "stop"),
		),
		ShowLog: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "log"),
		),
		Info: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "info"),
		),
		Start: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start"),
		),
	}
}

// A RefreshableItem holds the information about an instance that is refreshed constantly while initializing/
// stopping or exiting
type RefreshableItem struct {
	index int
	item  InstanceItem
}

var _ tea.Model = MainModel{}

// MainModel is the tea.Model of the main instance list
type MainModel struct {
	// loadedItems says whether the initial instance load has been completed
	loadedItems bool
	// List is the list model
	List list.Model
	// spinner is the spinner model used during instance loading
	spinner spinner.Model
	// keys is the main key map
	keys *ApplicationKeyMap
	// BasePath is a list of CloudControl instance base paths
	BasePath []string
	// Adapter is the adapter used to connect to CloudControl instances
	Adapter adapters.BaseAdapter
	// ShowInfo tells whether the info screen is shown
	ShowInfo bool
	// InfoItem holds the instance currently used in ShowInfo or ShowLog
	InfoItem InstanceItem
	// Width is the width of the screen
	Width int
	// Height is the height of the screen
	Height int
	// ListDisabled tells whether the instance list should not be shown
	ListDisabled bool
	// RunningConfirm tells wheter a confirmation action is currently running
	RunningConfirm bool
	// Confirm is the confirmation model currently used
	Confirm list.Model
	// ConfirmPrompt holds the prompt for the confirmation action
	ConfirmPrompt string
	// ItemsToRefresh is a list of items that need constant refreshing because they are currently
	// starting, initializing, invalid or stopping
	ItemsToRefresh []RefreshableItem
	// ShowLog tells whether the log screen is shown
	ShowLog bool
	// LogViewer is the textarea.Model used for log screens
	LogViewer textarea.Model
}

// NewMainModel creates a new model for the main instance list view
func NewMainModel(adapter adapters.BaseAdapter, basePath []string, items []list.Item) tea.Model {
	listKeys := NewApplicationKeyMap()

	// Set up the default item controller
	itemDelegate := list.NewDefaultDelegate()
	itemDelegate.SetHeight(4)
	itemDelegate.Styles.SelectedTitle = internal.SelectedItemTitleStyle
	itemDelegate.Styles.SelectedDesc = internal.SelectedItemDescriptionStyle
	itemDelegate.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{{
			listKeys.Run,
			listKeys.OpenCCC,
			listKeys.Restart,
			listKeys.Info,
			listKeys.Stop,
		}}
	}

	// Set up the instance list model

	instanceList := list.New(items, itemDelegate, 0, 0)
	instanceList.Title = "CloudControl instance manager"
	instanceList.Styles.Title = internal.TitleStyle
	instanceList.StatusMessageLifetime = 10 * time.Second
	instanceList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.ToggleTitleBar,
			listKeys.ToggleStatusBar,
			listKeys.TogglePagination,
			listKeys.ToggleHelpMenu,
			listKeys.Refresh,
		}
	}
	instanceList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.Run,
			listKeys.Info,
			listKeys.Restart,

			listKeys.Refresh,
		}
	}

	// Set up the spinner model

	s := spinner.New()
	s.Spinner = spinner.Dot

	// Set up the confirmation item list controller

	confirmDelegate := list.NewDefaultDelegate()
	confirmDelegate.ShowDescription = false
	confirmDelegate.Styles.NormalTitle = lipgloss.NewStyle().Padding(0)
	confirmDelegate.Styles.SelectedTitle = lipgloss.NewStyle().Padding(0).Foreground(lipgloss.Color("#ffffff"))
	confirmList := list.New([]list.Item{}, confirmDelegate, 0, 2)
	confirmList.SetShowFilter(false)
	confirmList.SetShowHelp(false)
	confirmList.SetShowPagination(false)
	confirmList.SetShowStatusBar(false)
	confirmList.SetShowTitle(false)
	confirmList.DisableQuitKeybindings()
	confirmList.SetFilteringEnabled(false)

	return MainModel{
		loadedItems: false,
		Adapter:     adapter,
		List:        instanceList,
		spinner:     s,
		keys:        listKeys,
		BasePath:    basePath,
		Confirm:     confirmList,
		LogViewer:   textarea.New(),
	}
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		LoadInstances,
		RefreshTick(),
	)
}
