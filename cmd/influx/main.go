package main

import (
	"fmt"
	"os"

	influxdb "github.com/influxdata/influxdb-client"
	"github.com/influxdata/influxdb-client/cmd/influx/cli"
	flag "github.com/spf13/pflag"
)

func main() {
	c := cli.New()

	config := cli.Config{}
	fs := flag.NewFlagSet("InfluxDB shell", flag.ExitOnError)
	fs.StringVarP(&c.Client.Addr, "host", "H", "", fmt.Sprintf("InfluxDB address [default: \"%s\"]", influxdb.DefaultAddr))
	fs.StringP("username", "u", "", "Username to connect to the server")
	fs.StringP("password", "p", "", "Password to connect to the server")
	fs.StringP("database", "d", "", "Default database for writes and queries")
	fs.StringVarP(&config.Format, "format", "f", "column", "Query format")
	fs.BoolP("version", "v", false, "Print the InfluxDB version and exit")
	fs.Parse(os.Args[1:])

	if err := c.Run(config); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
