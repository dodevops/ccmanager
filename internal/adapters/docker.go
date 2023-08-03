package adapters

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/cli"
	composeTypes "github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/cmd/formatter"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/go-resty/resty/v2"
	"github.com/moby/term"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var _ BaseAdapter = &DockerAdapter{}

// DockerAdapter implements CCmanager with docker and docker compose
type DockerAdapter struct {
	// dockerCLI holds the connection to a docker client
	dockerCLI *client.Client
	// composeBackend holds the connection to the docker compose service
	composeBackend *api.Service
}

func (d DockerAdapter) GetContainerStatus(basePath string, name string) (CloudControlStatus, error) {
	containerName := fmt.Sprintf("%s_cli_1", name)
	c := d.getClient()
	notFound := regexp.MustCompile("No such container")
	if i, err := c.ContainerInspect(context.Background(), containerName); err != nil {
		if notFound.Match([]byte(err.Error())) {
			return d.getContainerStatusFromCompose(basePath, name)
		}
		return CloudControlStatus{Error: err}, fmt.Errorf("can not inspect container %s: %w", name, err)
	} else {
		p := "n/a"
		cs := CCCUndef
		var err error
		if i.State.Running {
			if len(i.NetworkSettings.Ports["8080/tcp"]) == 1 {
				p = i.NetworkSettings.Ports["8080/tcp"][0].HostPort
				if i.State != nil && i.State.Running {
					cs, err = d.getCCCStatus(p)
				}
			} else {
				cs = CCCErr
				err = fmt.Errorf("CCC port not found or invalid")
			}
		} else if i.State.ExitCode != 0 {
			cs = CCCExited
			err = fmt.Errorf("container status: %s (Exit Code %d) %s", i.State.Status, i.State.ExitCode, i.State.Error)
		}
		var portMappings []PortMap

		for port, bindings := range i.NetworkSettings.Ports {
			if port != "8080/tcp" {
				portMappings = append(portMappings, PortMap{
					ContainerPort: port.Port(),
					HostPort:      bindings[0].HostPort,
				})
			}
		}
		return CloudControlStatus{
			Error:        err,
			Running:      i.State != nil && i.State.Running,
			Image:        strings.Split(i.Config.Image, ":")[0],
			Tag:          strings.Split(i.Config.Image, ":")[1],
			CCCPort:      p,
			CCCStatus:    cs,
			PortMappings: portMappings,
		}, nil
	}
}

func (d DockerAdapter) RunCloudControl(_ string, name string, consoleWidth uint, consoleHeight uint) (*ContainerExec, error) {
	containerName := fmt.Sprintf("%s_cli_1", name)
	consoleSize := [2]uint{consoleHeight, consoleWidth}
	return &ContainerExec{
		exec: func(stdin io.Reader, stdout io.Writer) error {
			quitWriter := make(chan bool)
			quitReader := make(chan bool)
			quitTermResize := make(chan bool)
			dockerCli := d.getClient()
			var executeID string
			if idResponse, err := dockerCli.ContainerExecCreate(context.Background(), containerName, types.ExecConfig{
				AttachStdout: true,
				AttachStderr: true,
				AttachStdin:  true,
				Tty:          true,
				ConsoleSize:  &consoleSize,
				Cmd:          []string{"/usr/local/bin/cloudcontrol", "run"},
			}); err != nil {
				return fmt.Errorf("can not create exec in container %s: %w", containerName, err)
			} else {
				executeID = idResponse.ID
			}

			var execResponse types.HijackedResponse
			if resp, err := dockerCli.ContainerExecAttach(
				context.Background(),
				executeID,
				types.ExecStartCheck{Tty: true, ConsoleSize: &consoleSize},
			); err != nil {
				return fmt.Errorf("can not attach to exec in container %s: %w", containerName, err)
			} else {
				execResponse = resp
				defer execResponse.Close()
			}
			fd, _ := term.GetFdInfo(stdin)
			var originalState *term.State
			if s, err := term.SetRawTerminal(fd); err != nil {
				return fmt.Errorf("can not set terminal to raw: %w", err)
			} else {
				originalState = s
			}

			if err := dockerCli.ContainerExecResize(context.Background(), executeID, types.ResizeOptions{
				Width:  consoleWidth,
				Height: consoleHeight,
			}); err != nil {
				return fmt.Errorf("can not resize tty of exec in container %s: %w", containerName, err)
			}

			go func(c net.Conn) {
				s := bufio.NewReader(c)
				for {
					select {
					case <-quitWriter:
						return
					default:
						if b, err := s.ReadByte(); err == nil {
							_, _ = stdout.Write([]byte{b})
						}
					}
				}
			}(execResponse.Conn)

			go func(c net.Conn) {
				s := bufio.NewReader(stdin)
				for {
					select {
					case <-quitReader:
						return
					default:
						if b, err := s.ReadByte(); err == nil {
							_, _ = c.Write([]byte{b})
						}
					}
				}
			}(execResponse.Conn)

			go func() {
				width := consoleWidth
				height := consoleHeight
				for {
					select {
					case <-quitTermResize:
						return
					default:
						if w, err := term.GetWinsize(fd); err == nil {
							if w.Width != uint16(width) || w.Height != uint16(height) {
								width = uint(w.Width)
								height = uint(w.Height)
								_ = dockerCli.ContainerExecResize(context.Background(), executeID, types.ResizeOptions{
									Width:  width,
									Height: height,
								})
							}
						}
					}
				}
			}()

			for {
				if execInspect, err := dockerCli.ContainerExecInspect(context.Background(), executeID); err != nil {
					return fmt.Errorf("error running in container %s: %w", containerName, err)
				} else {
					if execInspect.Running {
						continue
					}
					print("CloudControl closed. Press any key to proceed.")
					quitWriter <- true
					quitReader <- true
					quitTermResize <- true
					if err := term.RestoreTerminal(fd, originalState); err != nil {
						return fmt.Errorf("can not restore terminal: %w", err)
					}
					if execInspect.ExitCode != 0 {
						return fmt.Errorf("error running CloudControl in container %s: %s", containerName, "")
					}
					return nil
				}
			}
		},
	}, nil
}

func (d DockerAdapter) StartCloudControl(basePath string, name string) error {
	return d.up(basePath, name, true)
}

func (d DockerAdapter) StopCloudControl(basePath string, name string, _ bool) error {
	return d.down(basePath, name)
}
func (d DockerAdapter) GetLogs(basePath string, name string) (string, error) {
	var project *composeTypes.Project

	if p, err := d.getProject(basePath, name); err != nil {
		return "", err
	} else {
		project = p
	}
	c := d.getComposeBackend()
	lS := bytes.NewBufferString("")
	lC := formatter.NewLogConsumer(context.Background(), lS, lS, false, false, true)
	if err := c.Logs(context.Background(), project.Name, lC, api.LogOptions{
		Project: project,
	}); err != nil {
		return "", err
	}
	return lS.String(), nil
}

// getClient returns an already created connection to the docker client or creates one
func (d DockerAdapter) getClient() client.Client {
	if d.dockerCLI == nil {
		if dockerCli, err := client.NewClientWithOpts(client.FromEnv); err != nil {
			panic(fmt.Sprintf("Can not connect to Docker API: %s", err.Error()))
		} else {
			dockerCli.NegotiateAPIVersion(context.Background())
			d.dockerCLI = dockerCli
		}
	}
	return *d.dockerCLI
}

// getComposeBackend retuns an already open connection to the docker compose service or
// creates one
func (d DockerAdapter) getComposeBackend() api.Service {
	if d.composeBackend == nil {
		cl := d.getClient()
		// ensure old docker-compose compatibility
		api.Separator = "_"
		if c, err := command.NewDockerCli(command.WithAPIClient(&cl), command.WithDefaultContextStoreConfig()); err != nil {
			panic(fmt.Sprintf("Can not connect to Docker API: %s", err.Error()))
		} else {
			if err := c.Initialize(flags.NewClientOptions()); err != nil {
				panic(fmt.Sprintf("Can not initialize docker cli: %s", err.Error()))
			}
			s := compose.NewComposeService(c)
			d.composeBackend = &s
		}

	}
	return *d.composeBackend
}

// getCCCStatus retrieves status information as a CCCStatus struct from CCC listening on the given port
func (d DockerAdapter) getCCCStatus(port string) (CCCStatus, error) {
	type cccBackendStatus struct {
		Status string
	}
	c := resty.New()
	statusResult := cccBackendStatus{}
	if resp, err := c.R().SetResult(&statusResult).Get(fmt.Sprintf("http://localhost:%s/api/status", port)); err != nil {
		return CCCErr, err
	} else {
		if resp.IsError() {
			return CCCErr, nil
		}
		switch statusResult.Status {
		case "INIT":
			return CCCInit, nil
		case "INITIALIZED":
			return CCCReady, nil
		default:
			return CCCErr, fmt.Errorf("unknown CCC state: %s", statusResult.Status)
		}
	}
}

// getContainerStatusFromCompose is used if not enough information can be resolved from a running container
func (d DockerAdapter) getContainerStatusFromCompose(basePath string, name string) (CloudControlStatus, error) {
	var project *composeTypes.Project

	if p, err := d.getProject(basePath, name); err != nil {
		return CloudControlStatus{Error: err}, err
	} else {
		project = p
	}

	var cliService composeTypes.ServiceConfig
	if c, err := project.GetService("cli"); err != nil {
		return CloudControlStatus{Error: err}, err
	} else {
		cliService = c
	}
	image := strings.Split(cliService.Image, ":")[0]
	tag := strings.Split(cliService.Image, ":")[1]
	return CloudControlStatus{
		Running:   false,
		Image:     image,
		Tag:       tag,
		CCCStatus: CCCDown,
		CCCPort:   "n/a",
	}, nil
}

// getProject loads a docker compose project for an instance identified by basePath and name
func (d DockerAdapter) getProject(basePath string, name string) (*composeTypes.Project, error) {
	yamlFiles := []string{
		fmt.Sprintf("%s/docker-compose.yaml", filepath.Join(basePath, name)),
		fmt.Sprintf("%s/docker-compose.yml", filepath.Join(basePath, name)),
	}
	var yamlFile string
	for _, file := range yamlFiles {
		if _, err := os.Stat(file); err == nil {
			yamlFile = file
		}
	}

	if yamlFile == "" {
		err := fmt.Errorf("can't file docker-compose file for %s at path %s", name, basePath)
		return nil, err
	}

	var project *composeTypes.Project

	if p, err := cli.ProjectFromOptions(&cli.ProjectOptions{
		WorkingDir:  filepath.Join(basePath, name),
		ConfigPaths: []string{yamlFile},
		Environment: map[string]string{},
	}); err != nil {
		return nil, err
	} else {
		p.Name = name
		project = p
		for i, s := range project.Services {
			s.CustomLabels = map[string]string{
				api.ProjectLabel:     project.Name,
				api.ServiceLabel:     s.Name,
				api.VersionLabel:     api.ComposeVersion,
				api.WorkingDirLabel:  project.WorkingDir,
				api.ConfigFilesLabel: strings.Join(project.ComposeFiles, ","),
				api.OneoffLabel:      "False", // default, will be overridden by `run` command
			}
			project.Services[i] = s
		}
	}
	return project, nil
}

// up calls docker compose up on an instance
func (d DockerAdapter) up(path string, name string, pull bool) error {
	var project *composeTypes.Project
	if p, err := d.getProject(path, name); err != nil {
		return err
	} else {
		project = p
	}

	if pull {
		for _, service := range project.Services {
			service.PullPolicy = "Always"
		}
	}
	c := d.getComposeBackend()
	return c.Up(context.Background(), project, api.UpOptions{
		Create: api.CreateOptions{QuietPull: true, RemoveOrphans: true, Recreate: api.RecreateDiverged},
		Start:  api.StartOptions{Wait: true, Project: project},
	})
}

// down calls docker compose down on an instance
func (d DockerAdapter) down(path string, name string) error {
	var project *composeTypes.Project
	if p, err := d.getProject(path, name); err != nil {
		return err
	} else {
		project = p
	}
	c := d.getComposeBackend()
	return c.Down(context.Background(), project.Name, api.DownOptions{
		RemoveOrphans: true,
		Project:       project,
		Volumes:       true,
	})
}
