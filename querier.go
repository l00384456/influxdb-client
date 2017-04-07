package influxdb

import "io"

// QueryOptions is a set of configuration options for configuring queries.
type QueryOptions struct {
	Database  string
	Chunked   bool
	ChunkSize int
	Pretty    bool
	Format    string
	Async     bool
	Params    map[string]interface{}
}

// Clone creates a copy of the QueryOptions.
func (opt *QueryOptions) Clone() QueryOptions {
	clone := *opt
	clone.Params = make(map[string]interface{})
	for k, v := range opt.Params {
		clone.Params[k] = v
	}
	return clone
}

// Querier holds onto query options and acts as a convenience method for performing queries.
type Querier struct {
	c *Client
	QueryOptions
}

// Raw executes a raw query returns the unmodified io.ReadCloser from the
// response if a proper status code is returned.
func (q *Querier) Raw(query interface{}, opts ...QueryOption) (io.ReadCloser, string, error) {
	opt := q.QueryOptions
	if len(opts) > 0 {
		opt = opt.Clone()
		for _, f := range opts {
			f.apply(&opt)
		}
	}

	req, err := q.c.NewQueryRequest(query, opt)
	if err != nil {
		return nil, "", err
	}

	resp, err := q.c.Client.Do(req)
	if err != nil {
		return nil, "", err
	} else if resp.StatusCode/100 != 2 {
		return nil, "", ReadError(resp)
	}
	format := resp.Header.Get("Content-Type")
	return resp.Body, format, nil
}

// Select executes a query returns a Cursor that will parse the results from
// the stream. Use Execute for any queries that modify the database.
func (q *Querier) Select(query interface{}, opts ...QueryOption) (*Cursor, error) {
	querier := *q
	querier.Format = "json"
	r, format, err := querier.Raw(query, opts...)
	if err != nil {
		return nil, err
	}
	return NewCursor(r, format)
}

// Execute executes a query and returns if any error occurred. It discards the result.
func (q *Querier) Execute(query interface{}, opts ...QueryOption) error {
	cur, err := q.Select(query, opts...)
	if err != nil {
		return err
	}
	return cur.Each(func(*ResultSet) error { return nil })
}
