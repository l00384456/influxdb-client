package influxdb

import (
	"io"
	"time"
)

type Cursor struct {
	cur cursor
}

// NextSet will return the next ResultSet. This invalidates the previous
// ResultSet returned by this Cursor and discards any remaining data to be
// read (including any remaining partial results that need to be read).
// Depending on the implementation of the cursor, previous ResultSet's may
// still return results even after being invalidated.
func (c *Cursor) NextSet() (*ResultSet, error) { return c.cur.NextSet() }

// Close closes the cursor so the underlying stream will be closed if one exists.
func (c *Cursor) Close() error { return c.cur.Close() }

// Each iterates over every ResultSet in the Cursor.
func (c *Cursor) Each(fn func(*ResultSet) error) error {
	for {
		result, err := c.NextSet()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if err := fn(result); err != nil {
			if err == ErrStop {
				return nil
			}
			return err
		}
	}
}

// cursor is a cursor that reads and decodes a ResultSet.
type cursor interface {
	NextSet() (*ResultSet, error)
	Close() error
}

type ResultSet struct {
	result resultSet
}

func (r *ResultSet) Messages() []*Message         { return r.result.Messages() }
func (r *ResultSet) NextSeries() (*Series, error) { return r.result.NextSeries() }

// Each iterates over every Series in the ResultSet.
func (r *ResultSet) Each(fn func(*Series) error) error {
	for {
		series, err := r.NextSeries()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if err := fn(series); err != nil {
			if err == ErrStop {
				return nil
			}
			return err
		}
	}
}

// resultSet encapsulates a result from a single command.
type resultSet interface {
	// Messages returns the informational messages sent by the server for this ResultSet.
	Messages() []*Message

	// NextSeries returns the next series in the result.
	NextSeries() (*Series, error)
}

type Series struct {
	s series
}

func (s *Series) Name() string                { return s.s.Name() }
func (s *Series) Tags() Tags                  { return s.s.Tags() }
func (s *Series) Columns() []string           { return s.s.Columns() }
func (s *Series) Len() (n int, complete bool) { return s.s.Len() }
func (s *Series) NextRow() (Row, error)       { return s.s.NextRow() }

// Each iterates over every Row in the Series.
func (s *Series) Each(fn func(Row) error) error {
	for {
		row, err := s.s.NextRow()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if err := fn(row); err != nil {
			if err == ErrStop {
				return nil
			}
			return err
		}
	}
}

// series encapsulates a series within a ResultSet.
type series interface {
	// Name returns the measurement name associated with this series.
	Name() string

	// Tags returns the tags for this series. They are in sorted order.
	Tags() Tags

	// Columns returns the column names associated with this Series.
	Columns() []string

	// Len returns currently known length of the series. The length returned is
	// the cumulative length of the entire series, not just the current batch.
	// If the entire series hasn't been read because it is being sent in
	// partial chunks, this returns false for complete.
	Len() (n int, complete bool)

	// NextRow returns the next row in the result.
	NextRow() (Row, error)
}

// Row is a row of values in the ResultSet.
type Row interface {
	// Time returns the time column as a time.Time if it exists in the Row.
	Time() time.Time

	// Value returns value at index. If an invalid index is given, this will panic.
	Value(index int) interface{}

	// Values returns the values from the row as an array slice.
	Values() []interface{}

	// ValueByName returns the value by a named column. If the column does not
	// exist, this will return nil.
	ValueByName(column string) interface{}
}

// NewCursor constructs a new cursor from the io.ReadCloser and parses it with
// the appropriate decoder for the format. The following formatters are supported:
// msgpack (application/x-msgpack)
func NewCursor(r io.ReadCloser, format string) (*Cursor, error) {
	switch format {
	case "msgpack", "application/x-msgpack":
		cur, err := newMessagePackCursor(r)
		if err != nil {
			return nil, err
		}
		return &Cursor{cur: cur}, nil
	default:
		return nil, ErrUnknownFormat{Format: format}
	}
}

// Message is an informational message from the server.
type Message struct {
	Level string `json:"level"`
	Text  string `json:"text"`
}

func (m *Message) String() string {
	return m.Text
}
