# CCmanager ☁️ 🧰 📺 - CloudControl instance manager

If you're using [CloudControl](https://cloudcontrol.dodevops.io) to manage the infrastructure of multiple projects,
switching between different projects can be cumbersome.

CCmanager is a TUI for managing multiple CloudControl instances. Currently, it expects that all instances are managed
using `docker compose` in one or multiple subdirectories.

[![asciicast](https://asciinema.org/a/1sNBj2v0xJLAD7H4mqK1hGHEt.svg)](https://asciinema.org/a/1sNBj2v0xJLAD7H4mqK1hGHEt)

## Requirements

The current implementation only supports Docker with the Docker Compose v2 API.

## Usage

If you're placing all docker-compose files under directories like `$HOME/CloudControl/project1`,
`$HOME/CloudControl/project2`, you can set `CCMANAGER_BASEPATH` to `$HOME/CloudControl`.

If you're using multiple directories, `CCMANAGER_BASEPATH` supports a list of directories separated by `,`.

Afterwards, run CCmanager by placing the binary somewhere in your path and run

    ccmanager

CCmanager will load and show you the status of all found instances. Select an instance and use these keyboard
shortcuts:

- `enter`: Start a CloudControl shell
- `c`: Open CloudControlCenter, the CloudControl status web interface
- `L`: Show the log of the instance
- `r`: Restart the instance
- `d`: Stop the instance
- `s`: Start the instance
- `n`: Show an information screen about the instance

You can use `/` to filter the list of instances. For more shortcuts, press `h`.

## Container name separator

CCManager tries to lookup CloudControl environments by the name of their typical containers. These names are
usually in the form of <project><separator><service><separator><counter>.

The default separator is "-", but may be different in the container engine you're using. You can define the
separator using the environment variable `CCMANAGER_SEP`.

## Development

CCmanager is based on [Go](https://go.dev), 
[Bubbletea](https://pkg.go.dev/github.com/charmbracelet/bubbletea), 
[Bubbles](https://pkg.go.dev/github.com/charmbracelet/bubbles), 
[Lipgloss](https://pkg.go.dev/github.com/charmbracelet/lipgloss).

The Docker adapter is based on the [official Docker Client API](https://pkg.go.dev/github.com/docker/docker/client)
and [Docker Compose v2](https://pkg.go.dev/github.com/docker/compose/v2).

It adheres to the [Go standard project layout](https://github.com/golang-standards/project-layout).
