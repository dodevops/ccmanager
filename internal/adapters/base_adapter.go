package adapters

import (
	tea "github.com/charmbracelet/bubbletea"
	"io"
)

type PortMap struct {
	// ContainerPort holds the port inside the container
	ContainerPort string
	// HostPort holds the mapped port accessible from the host
	HostPort string
}

// CloudControlStatus holds various information about a cloudcontrol instance
type CloudControlStatus struct {
	// Error holds a possible error that has occured when gathering information about or running this instance
	Error error
	// Running says whether the instance is running
	Running bool
	// Image holds the image name the instance is using
	Image string
	// Tag holds the image tag the instance is using
	Tag string
	// CCCPort holds the port where the CCC can be reached
	CCCPort string
	// CCCStatus holds the status the CCC returns
	CCCStatus CCCStatus
	// PortMappings holds additional portmappings (aside from the CCCport)
	PortMappings []PortMap
}

// CCCStatus holds the instance status as returned by the CCC
type CCCStatus int64

const (
	// CCCUndef is an undefined status
	CCCUndef CCCStatus = iota
	// CCCDown describes an instance that has no containers
	CCCDown
	// CCCInit means that the instance is initializing
	CCCInit
	// CCCReady means that CloudControl can be executed in this instance
	CCCReady
	// CCCErr represents an error talking to the CCC
	CCCErr
	// CCCExited is used for containers that failed to start and exited
	CCCExited
)

// BaseAdapter describes the required functions to connect CCManager with container environments
type BaseAdapter interface {
	// GetContainerStatus fetches a CloudControlStatus from its backend. The instance is identified by
	// a basePath and the name of the instance
	GetContainerStatus(basePath string, name string) (CloudControlStatus, error)
	// RunCloudControl starts CloudControl in the given instance identified by basePath and
	// name. The consoleWidth and consoleHeight specify the width and height of the console window
	// that runs CloudControl. It returns a ContainerExec struct
	RunCloudControl(basePath string, name string, consoleWidth uint, consoleHeight uint) (*ContainerExec, error)
	// StartCloudControl starts the CloudControl instance identified by basePath and name
	StartCloudControl(basePath string, name string) error
	// StopCloudControl stops the CloudControl instance identified by basePath and name. The remove parameter
	// is set to true if the instance should be cleaned up after stopping
	StopCloudControl(basePath string, name string, remove bool) error
	// GetLogs returns the complete log of an instance identified by basePath and name
	GetLogs(basePath string, name string) (string, error)
}

// ContainerExec implements a tea.ExecCommand specialized for CCmanager
type ContainerExec struct {
	// stdin represents the reader to read input from
	stdin io.Reader
	// stdout represents the writer to write output to
	stdout io.Writer
	exec   func(reader io.Reader, writer io.Writer) error
}

var _ tea.ExecCommand = &ContainerExec{}

// Run executes the command. In this case the adapter starts CloudControl using
// stdin and stdout. StdErr is not used because we only use an interactive terminal
func (d *ContainerExec) Run() error {
	return d.exec(d.stdin, d.stdout)
}

func (d *ContainerExec) SetStdin(reader io.Reader) {
	d.stdin = reader
}

func (d *ContainerExec) SetStdout(writer io.Writer) {
	d.stdout = writer
}

// SetStderr intentionally does nothing. We don't use stderr.
func (d *ContainerExec) SetStderr(_ io.Writer) {
}
