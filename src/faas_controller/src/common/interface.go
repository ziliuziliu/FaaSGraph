package common

type GraphComputeRequest struct {
	Graph          string         `json:"graph"`
	RequestId      string         `json:"request_id"`
	App            string         `json:"app"`
	Function       string         `json:"function"`
	RoundLimit     int32          `json:"round_limit"`
	InEdge         bool           `json:"in_edge"`
	OutEdge        bool           `json:"out_edge"`
	AggregateOp    string         `json:"aggregate_op"`
	NoChange       bool           `json:"no_change"`
	TriggerAll     bool           `json:"trigger_all"`
	VertexArrayCnt int            `json:"vertex_array_cnt"`
	Param          map[string]any `json:"param"`
	GraphConfig    map[string]any
}

type AllocateContainerRequest struct {
	App      string `json:"app"`
	Function string `json:"function"`
	Key      string `json:"key"`
}

type RunResponse struct {
	NextActive []byte             `json:"activate"`
	Metrics    map[string]float64 `json:"metrics"`
}
