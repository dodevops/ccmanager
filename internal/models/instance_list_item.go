package models

import (
	"ccmanager/internal"
	"ccmanager/internal/adapters"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"path"
	"strings"
)

// InstanceItem is an implementation of list.Item that describes an CloudControl instance
// in the instance list
type InstanceItem struct {
	// Name is the instance name
	Name string
	// Path is the instance path
	Path string
	// State holds the instance status information in an adapters.CloudControlStatus struct
	State adapters.CloudControlStatus
}

var _ list.Item = InstanceItem{}

// Title shows the flavour and the name of the instance
func (i InstanceItem) Title() string {
	flavour := ""
	iParts := strings.Split(i.State.Image, "/")
	switch iParts[len(iParts)-1] {
	case "cloudcontrol-azure":
		flavour = internal.Labels["Azure"]
	case "cloudcontrol-gcloud":
		flavour = internal.Labels["GCP"]
	case "cloudcontrol-aws":
		flavour = internal.Labels["AWS"]
	case "cloudcontrol-tanzu":
		flavour = internal.Labels["Tanzu"]
	case "cloudcontrol-simple":
		flavour = internal.Labels["Simple"]
	default:
		if i.State.Error != nil {
			flavour = internal.Labels["Err"]
		} else {
			panic(fmt.Sprintf("%s unknown", i.State.Image))
		}
	}
	return fmt.Sprintf("%s %s", i.Name, flavour)
}

// Description holds the image, path and state of an instance
func (i InstanceItem) Description() string {
	return fmt.Sprintf(
		" %s:%s \n Path: %s \n State: %s",
		i.State.Image,
		i.State.Tag,
		path.Join(i.Path, i.Name),
		i.StateDescription(),
	)
}

// StateDescription describes the state of the instance
func (i InstanceItem) StateDescription() string {
	var s string
	if i.State.Error != nil && i.State.CCCStatus != adapters.CCCExited {
		s = fmt.Sprintf("System error: %s", i.State.Error)
	} else {
		switch i.State.CCCStatus {
		case adapters.CCCDown:
			s = "Down (use s to start)"
		case adapters.CCCInit:
			s = "Running/Initializing (use c to open CCC)"
		case adapters.CCCErr:
			s = "Running/Error"
		case adapters.CCCReady:
			s = "Running"
		case adapters.CCCExited:
			s = fmt.Sprintf("Container error: %s (use L to display the logs)", i.State.Error.Error())
		default:
			s = "Invalid"
		}
	}
	return s
}

// FilterValue only filters on instance name
func (i InstanceItem) FilterValue() string { return i.Name }
