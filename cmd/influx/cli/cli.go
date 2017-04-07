// Package cli contains the logic of the influx command line client.
package cli // import "github.com/influxdata/influxdb-client/cmd/influx/cli"

import (
	"fmt"

	influxdb "github.com/influxdata/influxdb-client"
	"github.com/peterh/liner"
)

// CommandLine holds CLI configuration and state.
type CommandLine struct {
	Line   *liner.State
	Client influxdb.Client
}

// New returns an instance of CommandLine with the specified client version.
func New() *CommandLine {
	return &CommandLine{}
}

// Run executes the CLI.
func (c *CommandLine) Run() error {
	serverInfo, err := c.Client.Ping()
	if err != nil {
		return err
	}

	addr := c.Client.Addr
	if addr == "" {
		addr = influxdb.DefaultAddr
	}
	fmt.Printf("Connected to %s version %s\n", addr, serverInfo.Version)
	return nil
}
