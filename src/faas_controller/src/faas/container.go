package faas

import (
	"application/src/common"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Container struct {
	containerType     string
	graph             string
	dataPath          string
	app               string
	function          string
	id                string
	ip                string
	port              int
	cpuSet            string
	workerIPList      []string
	containerAddrList []string
	startNodeList     []int32
}

func NewContainer(containerType string, graph string, dataPath string, app string, function string, ip string, port int, cpuSet string, workerIPList []string, containerAddrList []string, startNodeList []int32, fake bool) *Container {
	c := &Container{
		containerType:     containerType,
		graph:             graph,
		dataPath:          dataPath,
		app:               app,
		function:          function,
		ip:                ip,
		port:              port,
		cpuSet:            cpuSet,
		workerIPList:      workerIPList,
		containerAddrList: containerAddrList,
		startNodeList:     startNodeList,
	}
	if !fake {
		c.CreateContainer(containerType)
	}
	c.WaitStart()
	c.Init()
	return c
}

func (c *Container) GetId() string {
	return c.id
}

func (c *Container) GetIp() string {
	return c.ip
}

func (c *Container) GetPort() int {
	return c.port
}

func (c *Container) CreateContainer(containerType string) {
	req := map[string]any{"container_type": containerType, "app": c.app, "function": c.function, "port": c.port, "cpu_set": c.cpuSet, "data_path": c.dataPath}
	reqBody, _ := json.Marshal(req)
	createUrl := fmt.Sprintf("http://%s:20000/create_container", c.ip)
	rep, err := http.Post(createUrl, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		panic(err)
	}
	repBody, _ := ioutil.ReadAll(rep.Body)
	result := map[string]any{}
	json.Unmarshal(repBody, &result)
	c.id = result["id"].(string)
	rep.Body.Close()
}

func (c *Container) WaitStart() {
	statusUrl := fmt.Sprintf("http://%s:%d/status", c.ip, c.port)
	for {
		rep, err := http.Get(statusUrl)
		if err == nil && rep.StatusCode == 200 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (c *Container) Init() {
	req := map[string]any{"graph": c.graph, "ip": c.ip, "port": c.port, "app": c.app, "function": c.function, "config": common.Config, "worker_ip": c.workerIPList, "container_addr": c.containerAddrList, "start_node": c.startNodeList}
	reqBody, _ := json.Marshal(req)
	initUrl := fmt.Sprintf("http://%s:%d/init", c.ip, c.port)
	rep, _ := http.Post(initUrl, "application/json", bytes.NewReader(reqBody))
	rep.Body.Close()
}

func (c *Container) LoadKey(key string, sizes map[string]int32, inEdge bool, outEdge bool, aggregateOp string) {
	req := map[string]any{"key": key, "size": sizes[key], "in_edge": inEdge, "out_edge": outEdge, "aggregate_op": aggregateOp}
	reqBody, _ := json.Marshal(req)
	loadKeyUrl := fmt.Sprintf("http://%s:%d/load_key", c.ip, c.port)
	rep, err := http.Post(loadKeyUrl, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Println("Error container", c.ip, ":", c.port)
		panic(err)
	}
	rep.Body.Close()
}

func (c *Container) Destroy() {
	destroyUrl := fmt.Sprintf("http://%s:20000/remove_container", c.ip)
	req := map[string]string{"id": c.id}
	reqBody, _ := json.Marshal(req)
	rep, _ := http.Post(destroyUrl, "application/json", bytes.NewReader(reqBody))
	rep.Body.Close()
}
