package application

import (
	"application/src/common"
	"application/src/faas"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type GraphCompute struct {
	r           *common.GraphComputeRequest
	activeSet   *common.Bitmap
	round       int32
	start       float64
	performance map[string]float64
}

func GetTime() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}

func NewGraphCompute(r *common.GraphComputeRequest) *GraphCompute {
	c := &GraphCompute{
		r:           r,
		performance: make(map[string]float64),
	}
	return c
}

func (c *GraphCompute) Execute() (float64, float64, float64, float64, float64, float64, float64, float64) {
	c.start = GetTime()
	var containerSet *faas.ContainerSet
	newContainerTime, ioTime, sleep := 0.0, 0.0, 0.3
	for {
		containerSet, newContainerTime, ioTime = faas.Managers[c.r.Graph][fmt.Sprintf("%s-%s", c.r.App, c.r.Function)].GetContainerSet(c.r)
		if containerSet == nil {
			time.Sleep(time.Duration(sleep * float64(time.Second)))
			sleep *= 2
		} else {
			break
		}
	}
	preprocessTime, computeTime, flushTime, comm, maxComm, minComm := c.Run(containerSet)
	faas.Managers[c.r.Graph][fmt.Sprintf("%s-%s", c.r.App, c.r.Function)].PutContainerSet(containerSet)
	return newContainerTime, ioTime, preprocessTime, computeTime, flushTime, comm, maxComm, minComm
}

func (c *GraphCompute) Run(containerSet *faas.ContainerSet) (float64, float64, float64, float64, float64, float64) {
	reqBody := c.Serialize(containerSet)
	runUrl := fmt.Sprintf("http://%s:%d/run", containerSet.Coordinator.GetIp(), containerSet.Coordinator.GetPort())
	rep, _ := http.Post(runUrl, "application/json", bytes.NewReader(reqBody))
	result, err := ioutil.ReadAll(rep.Body)
	if err != nil {
		panic(err)
	}
	rep.Body.Close()
	return c.Deserialize(result)
}

func (c *GraphCompute) Serialize(containerSet *faas.ContainerSet) []byte {
	m := make(map[string]any)
	m["request_id"] = c.r.RequestId
	m["worker_ip_list"] = containerSet.WorkerIPList
	m["round_limit"] = c.r.RoundLimit
	m["no_change"] = c.r.NoChange
	m["param"] = c.r.Param
	m["graph_config"] = c.r.GraphConfig
	m["trigger_all"] = c.r.TriggerAll
	if c.r.VertexArrayCnt != 0 {
		m["vertex_array_cnt"] = c.r.VertexArrayCnt
	} else {
		m["vertex_array_cnt"] = 1
	}
	keyContainer := make(map[string]map[string]any)
	for key, container := range containerSet.GetKeyContainer() {
		keyContainer[key] = map[string]any{"ip": container.GetIp(), "port": container.GetPort()}
	}
	m["key_container"] = keyContainer
	data, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return data
}

func (c *GraphCompute) Deserialize(data []byte) (float64, float64, float64, float64, float64, float64) {
	m := make(map[string]any)
	json.Unmarshal(data, &m)
	return m["preprocess"].(float64), m["query"].(float64), m["store"].(float64), m["comm"].(float64), m["maxc"].(float64), m["minc"].(float64)
}
