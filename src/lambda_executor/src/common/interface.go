package common

type InitRequest struct {
	Graph             string         `json:"graph"`
	Ip                string         `json:"ip"`
	Port              int32          `json:"port"`
	App               string         `json:"app"`
	Function          string         `json:"function"`
	Config            map[string]any `json:"config"`
	WorkerIPList      []string       `json:"worker_ip"`
	ContainerAddrList []string       `json:"container_addr"`
	StartNodeList     []int32        `json:"start_node"`
}

type ReAssignRequest struct {
	App      string `json:"app"`
	Function string `json:"function"`
}

type LoadKeyRequest struct {
	Key         string `json:"key"`
	Size        int32  `json:"size"`
	InEdge      bool   `json:"in_edge"`
	OutEdge     bool   `json:"out_edge"`
	AggregateOp string `json:"aggregate_op"`
}

type LoadMemPoolRequest struct {
	RequestId      string `json:"request_id"`
	Key            string `json:"key"`
	VertexArrayCnt int    `json:"vertex_array_cnt"`
}

type LoadMemRequest struct {
	RequestId string         `json:"request_id"`
	Key       string         `json:"key"`
	Param     map[string]any `json:"param"`
}

type UpgradeRequest struct {
	Key string `json:"key"`
}

type RunParams struct {
	StartNode    int32  `json:"start_node"`
	ActiveVertex []byte `json:"active_vertex"`
}

type RunRequest struct {
	RequestId string    `json:"request_id"`
	Params    RunParams `json:"params"`
	NoChange  bool      `json:"no_change"`
}
