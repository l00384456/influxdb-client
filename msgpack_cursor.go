package influxdb

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/tinylib/msgp/msgp"
)

type msgpackCursor struct {
	r   io.ReadCloser
	dec *msgp.Reader
	cur *msgpackResult
}

func newMessagePackCursor(r io.ReadCloser) (*msgpackCursor, error) {
	dec := msgp.NewReader(r)
	var header msgpackResponseHeader
	if err := header.DecodeMsg(dec); err != nil {
		return nil, err
	}

	if header.Error != nil {
		return nil, errors.New(*header.Error)
	}
	return &msgpackCursor{r: r, dec: dec}, nil
}

func (c *msgpackCursor) NextSet() (ResultSet, error) {
	if c.cur != nil {
		if err := c.cur.Discard(); err != nil {
			return nil, err
		}
		c.cur = nil
	}

	var header msgpackResultHeader
	if err := header.DecodeMsg(c.dec); err != nil {
		return nil, err
	}

	if header.Error != nil {
		return nil, errors.New(*header.Error)
	}
	result := &msgpackResult{
		ID:  header.ID,
		dec: c.dec,
	}
	if err := result.readChunkHeader(); err != nil {
		return nil, err
	}
	c.cur = result
	return result, nil
}

func (c *msgpackCursor) Close() error {
	if c.cur != nil {
		if err := c.cur.Discard(); err != nil {
			return err
		}
		c.cur = nil
	}
	return c.r.Close()
}

type msgpackResult struct {
	ID        int
	dec       *msgp.Reader
	series    *msgpackSeries
	remaining int
	done      bool
}

func (r *msgpackResult) Messages() []*Message {
	return nil
}

func (r *msgpackResult) readChunkHeader() error {
	if r.dec == nil {
		return io.EOF
	}

	sz, err := r.dec.ReadArrayHeader()
	if err != nil {
		return err
	} else if sz != 2 {
		return fmt.Errorf("expected array of size 2: got %d", sz)
	}

	if count, err := r.dec.ReadInt(); err != nil {
		return err
	} else {
		r.remaining = count
	}
	if partial, err := r.dec.ReadBool(); err != nil {
		return err
	} else {
		r.done = !partial
	}
	return nil
}

func (r *msgpackResult) NextSeries() (Series, error) {
	if r.dec == nil {
		return nil, io.EOF
	}

	if r.series != nil {
		if err := r.series.Discard(); err != nil {
			return nil, err
		}
		r.series = nil
	}

	for r.remaining == 0 {
		if r.done {
			return nil, io.EOF
		}
		if err := r.readChunkHeader(); err != nil {
			return nil, err
		}
	}

	var header msgpackSeriesHeader
	if err := header.DecodeMsg(r.dec); err != nil {
		return nil, err
	}
	r.remaining--

	if header.Error != nil {
		return nil, errors.New(*header.Error)
	}

	series := &msgpackSeries{
		columns: header.Columns,
		dec:     r.dec,
	}
	if header.Name != nil {
		series.name = *header.Name
	}
	if err := series.readChunkHeader(); err != nil {
		return nil, err
	}
	r.series = series
	return series, nil
}

func (r *msgpackResult) Discard() error {
	if r.dec == nil {
		return nil
	}

	for {
		if _, err := r.NextSeries(); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

type msgpackSeries struct {
	name      string
	columns   []string
	dec       *msgp.Reader
	remaining int
	done      bool
	err       error
}

func (s *msgpackSeries) Name() string {
	return s.name
}

func (s *msgpackSeries) Tags() Tags {
	return nil
}

func (s *msgpackSeries) Columns() []string {
	return s.columns
}

func (s *msgpackSeries) Len() (n int, complete bool) {
	return 0, false
}

func (s *msgpackSeries) readChunkHeader() error {
	if s.dec == nil {
		return io.EOF
	}

	sz, err := s.dec.ReadArrayHeader()
	if err != nil {
		return err
	} else if sz != 2 {
		return fmt.Errorf("expected array of size 2: got %d", sz)
	}

	if count, err := s.dec.ReadInt(); err != nil {
		return err
	} else {
		s.remaining = count
	}
	if partial, err := s.dec.ReadBool(); err != nil {
		return err
	} else {
		s.done = !partial
	}
	return nil
}

func (s *msgpackSeries) NextRow() (Row, error) {
	if s.dec == nil {
		return nil, io.EOF
	} else if s.err != nil {
		return nil, s.err
	}

	for s.remaining == 0 {
		if s.done {
			return nil, io.EOF
		}
		if err := s.readChunkHeader(); err != nil {
			return nil, err
		}
	}

	var row msgpackRowHeader
	if err := row.DecodeMsg(s.dec); err != nil {
		return nil, err
	}
	s.remaining--

	if row.Error != nil {
		return nil, errors.New(*row.Error)
	}

	// If the number of remaining rows is zero and the series is not done, read
	// the next chunk header so we can fill in the length of the series to
	// something other than zero. If an error is returned from this, store it
	// so it can be returned on the next call instead of immediately.
	if !s.done {
		if err := s.readChunkHeader(); err != nil {
			s.err = err
			s.done = true
		}
	}
	return &msgpackRow{
		values:  row.Values,
		columns: s.Columns(),
	}, nil
}

func (s *msgpackSeries) Discard() error {
	if s.dec == nil {
		return nil
	}

	for {
		for s.remaining > 0 {
			if err := s.dec.Skip(); err != nil {
				return err
			}
			s.remaining--
		}

		if _, err := s.NextRow(); err != nil {
			if err == io.EOF {
				s.dec = nil
				return nil
			}
			return err
		}
	}
}

type msgpackRow struct {
	values  []interface{}
	columns []string
}

func (s *msgpackRow) Time() time.Time {
	t := s.ValueByName("time")
	if t != nil {
		return t.(time.Time)
	}
	return time.Time{}
}

func (s *msgpackRow) Value(index int) interface{} {
	return s.values[index]
}

func (s *msgpackRow) Values() []interface{} {
	return s.values
}

func (s *msgpackRow) ValueByName(column string) interface{} {
	if len(s.columns) == 0 {
		return nil
	}

	for i, name := range s.columns {
		if name == column {
			return s.Value(i)
		}
	}
	return nil
}
