package faas

import (
	"application/src/common"
	"math/rand"
	"strconv"
	"sync"
)

type ContainerSet struct {
	id                string
	r                 *common.GraphComputeRequest
	functionInfo      common.FunctionInfo
	waitQueue         []chan bool
	keyContainer      map[string]*Container
	WorkerIPList      []string
	ContainerAddrList []string
	StartNodeList     []int32
	containers        map[string]*Container
	Coordinator       *Container
	LastUsed          float64
	Status            string
}

func NewContainerSet(r *common.GraphComputeRequest, functionInfo common.FunctionInfo) (*ContainerSet, []string, []int, []string, string, int) {
	s := &ContainerSet{
		id:           strconv.Itoa(rand.Int()),
		r:            r,
		functionInfo: functionInfo,
		keyContainer: make(map[string]*Container),
		containers:   make(map[string]*Container),
		Status:       "STARTING",
	}
	partition := int(s.r.GraphConfig["PARTITION"].(float64))
	flag, ip, port, cpuSet, coordinatorIp, coordinatorPort := FaaSCluster.Get(s.functionInfo.MaxSlotPerContainer, partition)
	if !flag {
		return nil, nil, nil, nil, "", 0
	}
	return s, ip, port, cpuSet, coordinatorIp, coordinatorPort
}

func (s *ContainerSet) GetKeyContainer() map[string]*Container {
	return s.keyContainer
}

func (s *ContainerSet) PutResource() {
	ip := make([]string, 0)
	port := make([]int, 0)
	cpuSet := make([]string, 0)
	for _, container := range s.containers {
		ip = append(ip, container.ip)
		port = append(port, container.port)
		cpuSet = append(cpuSet, container.cpuSet)
	}
	FaaSCluster.Put(s.functionInfo.MaxSlotPerContainer, ip, port, cpuSet, s.Coordinator.ip, s.Coordinator.port)
}

func (s *ContainerSet) GetQueueLength() int {
	return len(s.waitQueue)
}

func (s *ContainerSet) AddQueue() chan bool {
	ch := make(chan bool, 1)
	s.waitQueue = append(s.waitQueue, ch)
	return ch
}

func (s *ContainerSet) NextQueue() bool {
	if len(s.waitQueue) == 0 {
		return false
	}
	s.waitQueue[0] <- true
	s.waitQueue = s.waitQueue[1:]
	return true
}

func (s *ContainerSet) SetRequest(r *common.GraphComputeRequest) {
	s.r = r
	s.Status = "RUNNING"
}

func (s *ContainerSet) LoadGraph(ip []string, port []int, cpuSet []string, coordinatorIp string, coordinatorPort int) (float64, float64) {
	s.Status = "LOAD_GRAPH"
	s.LastUsed = common.GetTime()
	partition := int(s.r.GraphConfig["PARTITION"].(float64))
	startNodes := s.r.GraphConfig["START_NODES"].([]int32)
	sizes := s.r.GraphConfig["SIZES"].(map[string]int32)
	dataPath := s.r.GraphConfig["DATA_PATH"].(string)
	wg := sync.WaitGroup{}
	m := sync.Mutex{}
	ipCnt := make(map[string]int)
	for i := range ip {
		if _, ok := ipCnt[ip[i]]; !ok {
			s.WorkerIPList = append(s.WorkerIPList, ip[i])
		}
		ipCnt[ip[i]]++
	}
	for i := range ip {
		for j := 0; j < s.functionInfo.MaxSlotPerContainer; j++ {
			if i+j >= partition {
				break
			}
			s.ContainerAddrList = append(s.ContainerAddrList, ip[i]+":"+strconv.Itoa(port[i]))
		}
		i += s.functionInfo.MaxSlotPerContainer
	}
	containerList := make([]*Container, len(ip))
	newGraphContainerFunc := func(ip string, port int, cpuSet string, id int, fake bool) {
		container := NewContainer("graph", s.r.Graph, dataPath, s.functionInfo.App, s.functionInfo.Function, ip, port, cpuSet, s.WorkerIPList, s.ContainerAddrList, s.r.GraphConfig["START_NODES"].([]int32), fake)
		m.Lock()
		s.containers[container.id] = container
		containerList[id] = container
		m.Unlock()
		wg.Done()
	}
	newCoordinatorContainerFunc := func(ip string, port int) {
		s.Coordinator = NewContainer("coordinator", s.r.Graph, dataPath, s.functionInfo.App, s.functionInfo.Function, ip, port, "0-"+strconv.Itoa(int(common.Config["MAX_SLOT_PER_NODE"].(float64))-1), s.WorkerIPList, s.ContainerAddrList, s.r.GraphConfig["START_NODES"].([]int32), false)
		wg.Done()
	}
	newContainerStart := common.GetTime()
	for i := range ip {
		wg.Add(1)
		go newGraphContainerFunc(ip[i], port[i], cpuSet[i], i, false)
		// go newGraphContainerFunc(ip[i], port[i], cpuSet[i], i, i != 0)
	}
	wg.Add(1)
	go newCoordinatorContainerFunc(coordinatorIp, coordinatorPort)
	wg.Wait()
	newContainerTime := common.GetTime() - newContainerStart
	loadKeyFunc := func(container *Container, key string) {
		container.LoadKey(key, sizes, s.r.InEdge, s.r.OutEdge, s.r.AggregateOp)
		wg.Done()
	}
	ioStart := common.GetTime()
	i := 0
	for _, container := range containerList {
		for j := 0; j < s.functionInfo.MaxSlotPerContainer; j++ {
			if i+j >= partition {
				break
			}
			key := strconv.Itoa(int(startNodes[i+j]))
			s.keyContainer[key] = container
			wg.Add(1)
			go loadKeyFunc(container, key)
		}
		i += s.functionInfo.MaxSlotPerContainer
	}
	wg.Wait()
	ioTime := common.GetTime() - ioStart
	return newContainerTime, ioTime
}

func (s *ContainerSet) Destroy() {
	wg := sync.WaitGroup{}
	destroyFunc := func(container *Container) {
		container.Destroy()
		wg.Done()
	}
	for _, container := range s.containers {
		wg.Add(1)
		go destroyFunc(container)
	}
	wg.Add(1)
	go destroyFunc(s.Coordinator)
	wg.Wait()
}
