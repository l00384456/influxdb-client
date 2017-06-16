package influxdb

//go:generate msgp -io=true -marshal=false -unexported -o msgpack_decode.go

type msgpackResponseHeader struct {
	Results int     `msg:"results"`
	Error   *string `msg:"error"`
}

type msgpackResultHeader struct {
	ID    int     `msg:"id"`
	Error *string `msg:"error"`
}

type msgpackSeriesHeader struct {
	Name    *string           `msg:"name"`
	Tags    map[string]string `msg:"tags"`
	Columns []string          `msg:"columns"`
	Error   *string           `msg:"error"`
}

type msgpackRowHeader struct {
	Values []interface{} `msg:"values"`
	Error  *string       `msg:"error"`
}
