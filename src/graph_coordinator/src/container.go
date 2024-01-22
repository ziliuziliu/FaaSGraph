package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Container struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}

func (c *Container) loadMemPool(requestId string, key string, vertexArrayCnt int) {
	req := map[string]any{"request_id": requestId, "key": key, "vertex_array_cnt": vertexArrayCnt}
	reqBody, _ := json.Marshal(req)
	loadMemUrl := fmt.Sprintf("http://%s:%d/load_mem_pool", c.Ip, c.Port)
	rep, err := http.Post(loadMemUrl, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Println("Error for container", c.Ip, ":", c.Port)
		panic(err)
	}
	_, err = ioutil.ReadAll(rep.Body)
	if err != nil {
		fmt.Println("Error for container", c.Ip, ":", c.Port)
		panic(err)
	}
	rep.Body.Close()
}

func (c *Container) loadMem(requestId string, key string, param map[string]any) []byte {
	req := map[string]any{"request_id": requestId, "key": key, "param": param}
	reqBody, _ := json.Marshal(req)
	loadMemUrl := fmt.Sprintf("http://%s:%d/load_mem", c.Ip, c.Port)
	rep, err := http.Post(loadMemUrl, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Println("Error for container", c.Ip, ":", c.Port)
		panic(err)
	}
	repBody, err := ioutil.ReadAll(rep.Body)
	if err != nil {
		fmt.Println("Error for container", c.Ip, ":", c.Port)
		panic(err)
	}
	rep.Body.Close()
	return repBody
}

func (c *Container) sendRequest(requestId string, startNode int32, activeVertex []byte, noChange bool) ([]byte, map[string]float64) {
	runUrl := fmt.Sprintf("http://%s:%d/run", c.Ip, c.Port)
	req, _ := http.NewRequest("POST", runUrl, bytes.NewReader(activeVertex))
	req.Header.Add("Request-Id", requestId)
	req.Header.Add("Start-Node", strconv.Itoa(int(startNode)))
	req.Header.Add("No-Change", strconv.FormatBool(noChange))
	rep, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error for container", c.Ip, ":", c.Port)
		panic(err)
	}
	activate, err := ioutil.ReadAll(rep.Body)
	if err != nil {
		fmt.Println("Error for container", c.Ip, ":", c.Port)
		panic(err)
	}
	var metrics map[string]float64
	metricsJson := rep.Header.Get("Metrics")
	json.Unmarshal([]byte(metricsJson), &metrics)
	rep.Body.Close()
	return activate, metrics
}

func (c *Container) flush() {
	flushUrl := fmt.Sprintf("http://%s:%d/flush", c.Ip, c.Port)
	rep, err := http.Post(flushUrl, "application/json", nil)
	if err != nil {
		fmt.Println("Error for container", c.Ip, ":", c.Port)
		panic(err)
	}
	rep.Body.Close()
}
