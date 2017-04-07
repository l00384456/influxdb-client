// Package cli contains the logic of the influx command line client.
package cli // import "github.com/influxdata/influxdb-client/cmd/influx/cli"

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"text/tabwriter"

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
func (c *CommandLine) Run(config Config) error {
	serverInfo, err := c.Client.Ping()
	if err != nil {
		return err
	}

	addr := c.Client.Addr
	if addr == "" {
		addr = influxdb.DefaultAddr
	}
	fmt.Printf("Connected to %s version %s\n", addr, serverInfo.Version)

	c.Line = liner.NewLiner()
	defer c.Line.Close()

	c.Line.SetMultiLineMode(true)

	querier := c.Client.Querier()
	querier.Chunked = true
	querier.Format = config.Format
	for {
		l, e := c.Line.Prompt("> ")
		if e == io.EOF {
			return nil
		} else if e != nil {
			return e
		}
		c.exec(querier, l)
	}
	return nil
}

func (c *CommandLine) exec(querier *influxdb.Querier, l string) {
	switch querier.Format {
	case "column":
		signalCh := make(chan os.Signal, 4)
		signal.Notify(signalCh, os.Interrupt)
		defer signal.Stop(signalCh)

		cur, err := querier.Select(l)
		if err != nil {
			fmt.Printf("ERR: %s\n", err)
			return
		}
		defer cur.Close()

		c.writeColumns(cur, signalCh)
	default:
		r, _, err := querier.Raw(l)
		if err != nil {
			fmt.Printf("ERR: %s\n", err)
			return
		}
		defer r.Close()
		io.Copy(os.Stdout, r)
	}
}

func (c *CommandLine) writeColumns(cur influxdb.Cursor, signalCh <-chan os.Signal) {
	writer := new(tabwriter.Writer)
	writer.Init(os.Stdout, 0, 8, 1, ' ', 0)

	influxdb.EachResult(cur, func(r influxdb.ResultSet) error {
		select {
		case <-signalCh:
			return influxdb.ErrStop
		default:
		}
		return influxdb.EachSeries(r, func(series influxdb.Series) error {
			fmt.Fprintf(writer, "name: %s\n", series.Name())
			if tags := series.Tags(); len(tags) > 0 {
				fmt.Fprintf(writer, "tags: %s\n", tags)
			}

			columns := series.Columns()
			fmt.Fprintln(writer, strings.Join(columns, "\t"))
			for i, col := range columns {
				if i > 0 {
					fmt.Fprint(writer, "\t")
				}
				fmt.Fprint(writer, strings.Repeat("-", len(col)))
			}
			fmt.Fprintln(writer)

			select {
			case <-signalCh:
				return influxdb.ErrStop
			default:
			}
			defer writer.Flush()
			return influxdb.EachRow(series, func(row influxdb.Row) error {
				select {
				case <-signalCh:
					return influxdb.ErrStop
				default:
				}

				values := row.Values()
				for i, val := range values {
					if i > 0 {
						fmt.Fprint(writer, "\t")
					}
					switch v := val.(type) {
					case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
						fmt.Fprintf(writer, "%d", v)
					default:
						fmt.Fprintf(writer, "%v", v)
					}
				}
				fmt.Fprintln(writer)
				return nil
			})
		})
	})
}
