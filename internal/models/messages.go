package models

import (
	"ccmanager/internal"
	"ccmanager/internal/adapters"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/browser"
	funk "github.com/thoas/go-funk"
	"log"
	"os"
	"strconv"
	"time"
)

// A ConfirmMsg starts a yes/no confirmation screen with a prompt. Depending on the choice, a tea.Cmd is issued.
type ConfirmMsg struct {
	Prompt        string
	YesMessage    tea.Cmd
	NoMessage     tea.Cmd
	DefaultChoice bool
}

func ConfirmMsgCmd(prompt string, yes tea.Cmd, no tea.Cmd, defaultChoice bool) tea.Cmd {
	return func() tea.Msg {
		return ConfirmMsg{
			Prompt:        prompt,
			YesMessage:    yes,
			NoMessage:     no,
			DefaultChoice: defaultChoice,
		}
	}
}

// An EnableListMsg enables the display of the main instance list
type EnableListMsg struct{}

func EnableList() tea.Msg {
	return EnableListMsg{}
}

// A DisableListMsg disables the display of the main instance list
type DisableListMsg struct{}

func DisableList() tea.Msg {
	return DisableListMsg{}
}

// An InstancesLoadedMsg is sent when all instances have been loaded
type InstancesLoadedMsg struct{}

func InstancesLoaded() tea.Msg {
	return InstancesLoadedMsg{}
}

// A LoadInstanceMsg triggers the (re-)load of one specific instance
type LoadInstanceMsg struct {
	BasePath string
	Name     string
}

// The LoadInstanceHandler loads information about an instance and - if it is loading an instance that is in
// MainModel.ItemsToRefresh replaces the item in the instance list with the new information
func LoadInstanceHandler(m MainModel, basePath string, name string) (MainModel, tea.Cmd) {
	item := InstanceItem{
		Name: name,
		Path: basePath,
	}

	if s, err := m.Adapter.GetContainerStatus(basePath, name); err == nil {
		item.State = s
	} else {
		item.State = adapters.CloudControlStatus{Error: err}
	}

	r := funk.Find(m.ItemsToRefresh, func(e RefreshableItem) bool {
		return e.item.Name == item.Name && e.item.Path == item.Path
	})

	if r != nil {
		ri := r.(RefreshableItem)
		cmd := m.List.SetItem(ri.index, item)
		if cmd != nil {
			return m, cmd
		}
		return m, nil
	}

	m.List.InsertItem(len(m.List.Items()), item)
	return m, nil
}

// The LoadInstancesMsg triggers the loading of all instances.
type LoadInstancesMsg struct{}

func LoadInstances() tea.Msg {
	return LoadInstancesMsg{}
}

// The LoadInstancesHandler runs through all configured base paths and generates a LoadInstanceMsg for every found
// instance configuration.
func LoadInstancesHandler(m MainModel) (MainModel, tea.Cmd) {
	var seq []tea.Cmd
	for _, p := range m.BasePath {
		for _, item := range getSubFolders(p) {
			seq = append(seq, func(path string, name string) tea.Cmd {
				return func() tea.Msg {
					return LoadInstanceMsg{
						BasePath: path,
						Name:     name,
					}
				}
			}(p, item))
		}
	}
	seq = append(seq, InstancesLoaded)
	return m, tea.Sequence(seq...)
}

// The OpenCCCMsg triggers opening a browser to point at the CCC
type OpenCCCMsg struct{}

func OpenCCC() tea.Msg {
	return OpenCCCMsg{}
}

// The OpenCCCHandler lets the operating system open a browser pointing to the currently selected instance's CCC
func OpenCCCHandler(m MainModel) (MainModel, tea.Cmd) {
	if cccPort, err := strconv.Atoi(m.List.SelectedItem().(InstanceItem).State.CCCPort); err == nil {
		if err := browser.OpenURL(fmt.Sprintf("http://localhost:%d", cccPort)); err != nil {
			return m, m.List.NewStatusMessage(internal.ErrorMessageStyle(fmt.Sprintf("Can not run browser: %s", err.Error())))
		}
	}
	return m, nil
}

// The RefreshTickMsg is used to constantly (using tea.Tick) refresh a list of initializing/stopping/failing instances
type RefreshTickMsg time.Time

func RefreshTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return RefreshTickMsg(t)
	})
}

// The ReloadItemsMsg triggers reloading all instances
type ReloadItemsMsg struct{}

func ReloadItems() tea.Msg {
	return ReloadItemsMsg{}
}

// RestartMsg triggers restarting an instance
type RestartMsg struct{}

func Restart() tea.Msg {
	return RestartMsg{}
}

// The RunCloudControlMsg triggers running CloudControl
type RunCloudControlMsg struct{}

func RunCloudControl() tea.Msg {
	return RunCloudControlMsg{}
}

// The RunCloudControlHandler manages running CloudControl on the currently selected instance. It uses the
// adapters.BaseAdapter.RounCloudControl function to generate an adapters.ContainerExec command.
func RunCloudControlHandler(m MainModel) tea.Cmd {
	item := m.List.SelectedItem().(InstanceItem)

	var runCmds []tea.Cmd
	if !item.State.Running {
		runCmds = append(runCmds, Start)
	} else if item.State.CCCStatus != adapters.CCCReady {
		if item.State.CCCStatus != adapters.CCCInit {
			return ConfirmMsgCmd("Instance has an invalid state. Do you want to restart the instance?", tea.Sequence(Stop, Start), nil, true)
		} else {
			return ConfirmMsgCmd("Instance is initializing. Do you want to open the CloudControl Center?", OpenCCC, nil, true)
		}
	}

	runCmds = append(runCmds, DisableList)
	runCmds = append(runCmds, tea.ExitAltScreen)
	runCmds = append(runCmds, tea.ClearScreen)

	if c, err := m.Adapter.RunCloudControl(item.Path, item.Name, uint(m.Width), uint(m.Height)); err != nil {
		return m.List.NewStatusMessage(internal.ErrorMessageStyle(err.Error()))
	} else {
		runCmds = append(runCmds, tea.Exec(c, func(err error) tea.Msg {
			if err != nil {
				return m.List.NewStatusMessage(internal.ErrorMessageStyle(err.Error()))
			}
			return nil
		}))
	}

	runCmds = append(runCmds, tea.EnterAltScreen)
	runCmds = append(runCmds, EnableList)

	return tea.Sequence(runCmds...)
}

// ShowInfoMsg is used to show information about the currently selected instance.
type ShowInfoMsg struct{}

func ShowInfo() tea.Msg {
	return ShowInfoMsg{}
}

// ShowLogMsg is sent to show the log of an instance
type ShowLogMsg struct{}

func ShowLog() tea.Msg {
	return ShowLogMsg{}
}

// ShowLogHandler fetches the currently selected instance's log enables the logviewer
func ShowLogHandler(m MainModel) (MainModel, tea.Cmd) {
	item := m.List.SelectedItem().(InstanceItem)

	if l, err := m.Adapter.GetLogs(item.Path, item.Name); err != nil {
		m.LogViewer.SetContent(fmt.Sprintf("Error getting logs: %s", err.Error()))
	} else {
		m.LogViewer.SetContent(l)
	}

	m.ShowLog = true
	return m, tea.Sequence(
		tea.ClearScreen,
		DisableList,
	)
}

// StartMsg starts an instance
type StartMsg struct{}

func Start() tea.Msg {
	return StartMsg{}
}

// StartHandler uses adapters.BaseAdapter.StartCloudControl to start an instance
func StartHandler(m MainModel) tea.Cmd {
	item := m.List.SelectedItem().(InstanceItem)
	return func() tea.Msg {
		var cmds []tea.Cmd
		if err := m.Adapter.StartCloudControl(item.Path, item.Name); err != nil {
			cmds = append(cmds, m.List.NewStatusMessage(internal.ErrorMessageStyle(fmt.Sprintf("Can not start CloudControl: %s", err.Error()))))
		}
		if cmds != nil {
			return tea.Batch(cmds...)()
		}
		return nil
	}
}

// StopMsg is used to stop an instance.
type StopMsg struct{}

func Stop() tea.Msg {
	return StopMsg{}
}

// StopHandler uses adapters.BaseAdapter.StopCloudControl to stop an instance.
func StopHandler(m MainModel) tea.Cmd {
	item := m.List.SelectedItem().(InstanceItem)
	return func() tea.Msg {
		var cmds []tea.Cmd
		if err := m.Adapter.StopCloudControl(item.Path, item.Name, true); err != nil {
			cmds = append(cmds, m.List.NewStatusMessage(internal.ErrorMessageStyle(fmt.Sprintf("Can not start CloudControl: %s", err.Error()))))
		}
		if cmds != nil {
			return tea.Batch(cmds...)()
		}
		return nil
	}
}

// getSubFolders walks through the given BasePath and returns a list of directory entries which are also directories.
func getSubFolders(basePath string) []string {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		log.Fatal(err)
	}

	var items []string

	for _, e := range entries {
		if e.Type().IsDir() {
			items = append(items, e.Name())
		}
	}
	return items
}
