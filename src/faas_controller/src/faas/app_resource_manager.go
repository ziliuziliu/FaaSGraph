package faas

import (
	"application/src/common"
	"math"
	"sync"
	"time"
)

type AppResourceManager struct {
	functionInfo           common.FunctionInfo
	freeContainerPool      []*ContainerSet
	availableContainerPool map[string]*ContainerSet
	mutex                  sync.Mutex
	ticker                 *time.Ticker
}

const ALLOW_IDLE_TIME = 1200

var Managers map[string]map[string]*AppResourceManager

func init() {
	Managers = map[string]map[string]*AppResourceManager{}
	graphConfigs := common.Config["GRAPH"].(map[string]any)
	for graph := range graphConfigs {
		Managers[graph] = map[string]*AppResourceManager{}
		for name, info := range common.Info {
			Managers[graph][name] = NewAppResourceManager(info)
			go Managers[graph][name].Recycle()
		}
	}
}

func NewAppResourceManager(info common.FunctionInfo) *AppResourceManager {
	m := &AppResourceManager{
		functionInfo:           info,
		freeContainerPool:      make([]*ContainerSet, 0),
		availableContainerPool: make(map[string]*ContainerSet),
		mutex:                  sync.Mutex{},
		ticker:                 time.NewTicker(time.Second * 30),
	}
	return m
}

func (m *AppResourceManager) GetContainerSet(r *common.GraphComputeRequest) (*ContainerSet, float64, float64) {
	m.mutex.Lock()
	if len(m.freeContainerPool) > 0 {
		containerSet := m.freeContainerPool[len(m.freeContainerPool)-1]
		m.freeContainerPool = m.freeContainerPool[0 : len(m.freeContainerPool)-1]
		m.availableContainerPool[containerSet.id] = containerSet
		containerSet.LastUsed = common.GetTime()
		m.mutex.Unlock()
		containerSet.SetRequest(r)
		return containerSet, 0, 0
	} else if int(common.Config["QUEUE_LENGTH"].(float64)) != 0 && len(m.availableContainerPool) > 0 {
		queue_length_limit := int(common.Config["QUEUE_LENGTH"].(float64))
		containerSetId := "NEW"
		shortest_length := math.MaxInt
		for id, containerSet := range m.availableContainerPool {
			if length := containerSet.GetQueueLength(); length < queue_length_limit && length < shortest_length {
				shortest_length = length
				containerSetId = id
			}
		}
		if containerSetId != "NEW" {
			containerSet := m.availableContainerPool[containerSetId]
			ch := containerSet.AddQueue()
			m.mutex.Unlock()
			<-ch
			containerSet.SetRequest(r)
			return containerSet, 0, 0
		}
	}
	containerSet, ip, port, cpuSet, coordinatorIp, coordinatorPort := NewContainerSet(r, m.functionInfo)
	if containerSet == nil {
		m.mutex.Unlock()
		return nil, 0, 0
	} else { // successfully allocated
		m.availableContainerPool[containerSet.id] = containerSet
		m.mutex.Unlock()
		newContainerTime, ioTime := containerSet.LoadGraph(ip, port, cpuSet, coordinatorIp, coordinatorPort)
		containerSet.SetRequest(r)
		return containerSet, newContainerTime, ioTime
	}
}

func (m *AppResourceManager) PutContainerSet(containerSet *ContainerSet) {
	m.mutex.Lock()
	if !containerSet.NextQueue() {
		delete(m.availableContainerPool, containerSet.id)
		m.freeContainerPool = append(m.freeContainerPool, containerSet)
	}
	m.mutex.Unlock()
}

func (m *AppResourceManager) Recycle() {
	for {
		<-m.ticker.C
		cur := common.GetTime()
		remove := make([]*ContainerSet, 0)
		newContainerPool := make([]*ContainerSet, 0)
		m.mutex.Lock()
		for _, containerSet := range m.freeContainerPool {
			if cur-containerSet.LastUsed > ALLOW_IDLE_TIME {
				remove = append(remove, containerSet)
			} else {
				newContainerPool = append(newContainerPool, containerSet)
			}
		}
		m.freeContainerPool = newContainerPool
		m.mutex.Unlock()
		for _, containerSet := range remove {
			containerSet.Destroy()
			containerSet.PutResource()
		}
	}
}
