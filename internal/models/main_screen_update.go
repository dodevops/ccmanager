package models

import (
	"ccmanager/internal"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Update holds the main controlling code for the applcation.
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// If the info screen is shown, only react to a keypress and hide it.
	if m.ShowInfo {
		switch msg.(type) {
		case tea.KeyMsg:
			m.ShowInfo = false
			return m, nil
		}
	}

	// If the log view is shown, only react to the quit keys and update the logviewer model.
	if m.ShowLog {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if key.Matches(msg, m.List.KeyMap.Quit) {
				m.ShowLog = false
				return m, nil
			}
		}
		newLogViewerModel, cmd := m.LogViewer.Update(msg)
		m.LogViewer = newLogViewerModel
		return m, cmd
	}

	// If the confirm view is shown, only react to the selection and update the confirmation model.
	if m.RunningConfirm {
		switch msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg.(tea.KeyMsg), m.keys.Run):
				m.RunningConfirm = false
				return m, tea.Sequence(
					EnableList,
					m.Confirm.SelectedItem().(ConfirmItem).command,
				)
			}
		}
		newConfirmModel, cmd := m.Confirm.Update(msg)
		m.Confirm = newConfirmModel
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := internal.AppStyle.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v)
		m.LogViewer.SetWidth(msg.Width)
		m.LogViewer.SetHeight(msg.Height - 2)
		m.Width = msg.Width
		m.Height = msg.Height
		m.List.Styles.Title.Width(m.Width - h - 4)

	case tea.KeyMsg:
		return m.keyHandler(msg)

	case LoadInstancesMsg:
		return LoadInstancesHandler(m)
	case LoadInstanceMsg:
		return LoadInstanceHandler(m, msg.BasePath, msg.Name)
	case InstancesLoadedMsg:
		m.loadedItems = true
		return m, nil
	case ReloadItemsMsg:
		for len(m.List.Items()) > 0 {
			m.List.RemoveItem(0)
		}
		m.ItemsToRefresh = nil
		return m, LoadInstances

	case ShowInfoMsg:
		m.ShowInfo = true
		m.InfoItem = m.List.SelectedItem().(InstanceItem)
	case OpenCCCMsg:
		return OpenCCCHandler(m)
	case RunCloudControlMsg:
		return m, RunCloudControlHandler(m)
	case StartMsg:
		return m, tea.Sequence(DisableList, tea.ClearScreen, StartHandler(m), tea.ClearScreen, EnableList)
	case StopMsg:
		return m, tea.Sequence(DisableList, tea.ClearScreen, StopHandler(m), tea.ClearScreen, EnableList)
	case RestartMsg:
		return m, tea.Sequence(DisableList, tea.ClearScreen, StopHandler(m), StartHandler(m), tea.ClearScreen, EnableList)
	case ShowLogMsg:
		return ShowLogHandler(m)

	case RefreshTickMsg:
		if m.loadedItems {
			var refreshCmds []tea.Cmd
			m.ItemsToRefresh = nil
			for index, item := range m.List.Items() {
				m.ItemsToRefresh = append(m.ItemsToRefresh, RefreshableItem{
					index,
					item.(InstanceItem),
				})
			}
			for _, itemToRefresh := range m.ItemsToRefresh {
				refreshCmds = append(refreshCmds, func(i RefreshableItem) tea.Cmd {
					return func() tea.Msg {
						return LoadInstanceMsg{
							BasePath: i.item.Path,
							Name:     i.item.Name,
						}
					}
				}(itemToRefresh))
			}

			return m, tea.Sequence(
				tea.Batch(refreshCmds...),
				RefreshTick(),
			)
		}

	case EnableListMsg:
		m.ListDisabled = false
		return m, nil
	case DisableListMsg:
		m.ListDisabled = true
		return m, nil

	case ConfirmMsg:
		m.RunningConfirm = true
		m.ConfirmPrompt = msg.Prompt
		for len(m.Confirm.Items()) > 0 {
			m.Confirm.RemoveItem(0)
		}
		m.Confirm.InsertItem(0, ConfirmItem{
			title:   "Yes",
			command: msg.YesMessage,
		})
		m.Confirm.InsertItem(1, ConfirmItem{
			title:   "No",
			command: msg.NoMessage,
		})
		if !msg.DefaultChoice {
			m.Confirm.Select(1)
		}
		return m, DisableList

	default:
		if !m.loadedItems {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	newListModel, cmd := m.List.Update(msg)
	m.List = newListModel
	return m, cmd
}

// The keyHandler reacts to key presses while the main screen is shown.
func (m MainModel) keyHandler(msg tea.KeyMsg) (MainModel, tea.Cmd) {
	switch {
	case m.List.FilterState() == list.Filtering:
		break

	case key.Matches(msg, m.keys.ToggleTitleBar):
		v := !m.List.ShowTitle()
		m.List.SetShowTitle(v)
		m.List.SetShowFilter(v)
		m.List.SetFilteringEnabled(v)
		return m, nil

	case key.Matches(msg, m.keys.ToggleStatusBar):
		m.List.SetShowStatusBar(!m.List.ShowStatusBar())
		return m, nil

	case key.Matches(msg, m.keys.TogglePagination):
		m.List.SetShowPagination(!m.List.ShowPagination())
		return m, nil

	case key.Matches(msg, m.keys.ToggleHelpMenu):
		m.List.SetShowHelp(!m.List.ShowHelp())
		return m, nil

	case key.Matches(msg, m.keys.Refresh):
		return m, ReloadItems
	case key.Matches(msg, m.keys.Run):
		return m, RunCloudControl
	case key.Matches(msg, m.keys.Info):
		return m, ShowInfo
	case key.Matches(msg, m.keys.OpenCCC):
		return m, OpenCCC
	case key.Matches(msg, m.keys.Stop):
		return m, Stop
	case key.Matches(msg, m.keys.Start):
		return m, Start
	case key.Matches(msg, m.keys.Restart):
		return m, Restart
	case key.Matches(msg, m.keys.ShowLog):
		return m, ShowLog
	}
	if m.ShowLog {
		newLogViewerModel, cmd := m.LogViewer.Update(msg)
		m.LogViewer = newLogViewerModel
		return m, cmd
	} else {
		newListModel, cmd := m.List.Update(msg)
		m.List = newListModel
		return m, cmd
	}
}
