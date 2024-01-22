package main

import (
	"encoding/binary"
	"fmt"
	"graph_coordinator/src/common"
	"math"
	"net/http"
	"strconv"
	"sync"
)

type ContainerSet struct {
	RequestId      string               `json:"request_id"`
	KeyContainer   map[string]Container `json:"key_container"`
	WorkerIPList   []string             `json:"worker_ip_list"`
	RoundLimit     int                  `json:"round_limit"`
	NoChange       bool                 `json:"no_change"`
	TriggerAll     bool                 `json:"trigger_all"`
	VertexArrayCnt int                  `json:"vertex_array_cnt"`
	Param          map[string]any       `json:"param"`
	GraphConfig    map[string]any       `json:"graph_config"`
}

func (s *ContainerSet) run() (float64, float64, float64, int64, int64, int64) {
	// start := common.GetTime()
	totalNode := int(s.GraphConfig["TOTAL_NODE"].(float64))
	preprocessStart := common.GetTime()
	for _, workerIp := range s.WorkerIPList {
		s.allocateMem(workerIp)
	}
	s.loadMemPool(s.RequestId)
	activeSet, round, perf := s.loadMem(s.RequestId, s.Param), 0, make(map[string]float64)
	preprocessTime := common.GetTime() - preprocessStart
	computeStart := common.GetTime()
	for !activeSet.Empty() {
		round++
		roundStart := common.GetTime()
		if s.RoundLimit > 0 && round > s.RoundLimit {
			break
		}
		fmt.Println("active", activeSet.Count())
		if s.TriggerAll {
			activeSet = common.NewBitmapWithAll(totalNode)
		}
		active, performance, runOverall, _ := s.dispatchRequest(activeSet)
		if !s.NoChange {
			activeSet = active
		}
		fmt.Println("round", round, "last round", common.GetTime()-roundStart)
		fmt.Println("metrics", performance)
		fmt.Println(runOverall)
		s.combinePerformanceMetrics(perf, performance)
	}
	computeTime := common.GetTime() - computeStart
	fmt.Println("query", computeTime)
	flushStart := common.GetTime()
	s.flush()
	flushTime := common.GetTime() - flushStart
	comm, maxComm, minComm := int64(0), int64(0), int64(1000000000000000)
	for _, workerIp := range s.WorkerIPList {
		localComm := s.clearMem(workerIp)
		comm += localComm
		if localComm > maxComm {
			maxComm = localComm
		}
		if localComm < minComm {
			minComm = localComm
		}
	}
	return preprocessTime, computeTime, flushTime, comm, maxComm, minComm
}

func (s *ContainerSet) allocateMem(workerIp string) {
	allocateMemUrl := fmt.Sprintf("http://%s:20000/allocate_mem", workerIp)
	req, _ := http.NewRequest("POST", allocateMemUrl, nil)
	req.Header.Add("Request-Id", s.RequestId)
	req.Header.Add("Total-Node", strconv.Itoa(int(s.GraphConfig["TOTAL_NODE"].(float64))))
	req.Header.Add("Vertex-Array-Cnt", strconv.Itoa(s.VertexArrayCnt))
	rep, _ := http.DefaultClient.Do(req)
	rep.Body.Close()
}

func (s *ContainerSet) clearMem(workerIp string) int64 {
	clearMemUrl := fmt.Sprintf("http://%s:20000/clear_mem", workerIp)
	req, _ := http.NewRequest("POST", clearMemUrl, nil)
	req.Header.Add("Request-Id", s.RequestId)
	rep, _ := http.DefaultClient.Do(req)
	b := make([]byte, 8)
	rep.Body.Read(b)
	rep.Body.Close()
	return int64(binary.LittleEndian.Uint64(b))
}

func (s *ContainerSet) loadMemPool(requestId string) {
	startNodes := s.GraphConfig["START_NODES"].([]any)
	wg := sync.WaitGroup{}
	loadMemPool := func(key string, container Container) {
		container.loadMemPool(requestId, key, s.VertexArrayCnt)
		wg.Done()
	}
	for _, startNode := range startNodes {
		key := strconv.Itoa(int(startNode.(float64)))
		wg.Add(1)
		go loadMemPool(key, s.KeyContainer[key])
	}
	wg.Wait()
}

func (s *ContainerSet) loadMem(requestId string, param map[string]any) *common.Bitmap {
	totalNode := int(s.GraphConfig["TOTAL_NODE"].(float64))
	startNodes := s.GraphConfig["START_NODES"].([]any)
	wg := sync.WaitGroup{}
	activate := common.NewBitmap(totalNode)
	loadMem := func(key string, container Container) {
		buf := container.loadMem(requestId, key, param)
		bitmap := common.NewBitmapWith(common.ByteSlice2Uint32Slice(buf))
		activate.Or(bitmap)
		wg.Done()
	}
	for _, startNode := range startNodes {
		key := strconv.Itoa(int(startNode.(float64)))
		wg.Add(1)
		go loadMem(key, s.KeyContainer[key])
	}
	wg.Wait()
	return activate
}

func (s *ContainerSet) dispatchRequest(activeSet *common.Bitmap) (*common.Bitmap, map[string]float64, map[int32]float64, map[int32]float64) {
	totalNode := int(s.GraphConfig["TOTAL_NODE"].(float64))
	startNodes := s.GraphConfig["START_NODES"].([]any)
	sizes := s.GraphConfig["SIZES"].(map[string]any)
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	performance := make(map[string]float64)
	runOverall := make(map[int32]float64)
	gasRun := make(map[int32]float64)
	nextActiveSet := common.NewBitmap(totalNode)
	dispatch := func(startNode int32, size int32, container Container) {
		activate, metrics := container.sendRequest(s.RequestId, startNode, common.Uint32Slice2ByteSlice(activeSet.Data), s.NoChange)
		repBitmap := common.NewBitmapWith(common.ByteSlice2Uint32Slice(activate))
		if len(startNodes) <= 2 {
			mutex.Lock()
			nextActiveSet.Or2(repBitmap)
			for metric, t := range metrics {
				if metric == "run_overall" {
					runOverall[startNode] = t
				} else if metric == "gas_run" {
					gasRun[startNode] = t
				}
				if _, ok := performance[metric]; !ok {
					performance[metric] = t
				} else if metric == "scanned_edge" {
					performance[metric] += t
				} else {
					performance[metric] = math.Max(performance[metric], t)
				}
			}
			mutex.Unlock()
		} else {
			nextActiveSet.Or(repBitmap)
			mutex.Lock()
			for metric, t := range metrics {
				if metric == "run_overall" {
					runOverall[startNode] = t
				} else if metric == "gas_run" {
					gasRun[startNode] = t
				}
				if _, ok := performance[metric]; !ok {
					performance[metric] = t
				} else if metric == "scanned_edge" {
					performance[metric] += t
				} else {
					performance[metric] = math.Max(performance[metric], t)
				}
			}
			mutex.Unlock()
		}
		wg.Done()
	}
	for _, startNode := range startNodes {
		key := strconv.Itoa(int(startNode.(float64)))
		size := int32(sizes[key].(float64))
		container := s.KeyContainer[key]
		wg.Add(1)
		go dispatch(int32(startNode.(float64)), size, container)
	}
	wg.Wait()
	return nextActiveSet, performance, runOverall, gasRun
}

func (s *ContainerSet) flush() {
	wg := sync.WaitGroup{}
	flushFunc := func(container Container) {
		container.flush()
		wg.Done()
	}
	for _, container := range s.KeyContainer {
		wg.Add(1)
		go flushFunc(container)
	}
	wg.Wait()
}

func (s *ContainerSet) combinePerformanceMetrics(a map[string]float64, b map[string]float64) {
	for metric, t := range b {
		if _, ok := a[metric]; !ok {
			a[metric] = t
		} else {
			a[metric] += t
		}
	}
}
