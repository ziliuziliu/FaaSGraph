package faas

import (
	"application/src/common"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	LOW_PORT_NUMBER  = 19000
	HIGH_PORT_NUMBER = 20000
)

type Cluster struct {
	cond        *sync.Cond
	nodes       []*Node
	currentSlot int
}

type Node struct {
	ip      string
	slot    int
	ports   []int
	cpuMask []bool
}

var FaaSCluster *Cluster

func init() {
	maxSlotPerNode := int(common.Config["MAX_SLOT_PER_NODE"].(float64))
	currentTakenSlotConfig := common.Config["CURRERNT_TAKEN_SLOT"].([]any)
	currentTakenSlot := make([]int, 0)
	totalTakenSlot := 0
	if len(currentTakenSlotConfig) == 0 {
		for range common.WorkerList {
			currentTakenSlot = append(currentTakenSlot, 0)
		}
	} else {
		for i := range common.WorkerList {
			currentTakenSlot = append(currentTakenSlot, int(currentTakenSlotConfig[i].(float64)))
			totalTakenSlot += currentTakenSlot[i]
		}
	}
	FaaSCluster = &Cluster{
		cond:        sync.NewCond(&sync.Mutex{}),
		nodes:       make([]*Node, 0),
		currentSlot: maxSlotPerNode*len(common.WorkerList) - totalTakenSlot,
	}
	for i, ip := range common.WorkerList {
		FaaSCluster.nodes = append(FaaSCluster.nodes, &Node{
			ip:      ip,
			slot:    maxSlotPerNode - currentTakenSlot[i],
			ports:   make([]int, 0),
			cpuMask: make([]bool, maxSlotPerNode),
		})
		for j := LOW_PORT_NUMBER; j <= HIGH_PORT_NUMBER; j++ {
			FaaSCluster.nodes[i].ports = append(FaaSCluster.nodes[i].ports, j)
		}
	}
}

func (c *Cluster) Get(slotPerContainer int, totalSlot int) (bool, []string, []int, []string, string, int) {
	if totalSlot < 2 {
		totalSlot = 2
	}
	shareCPU := common.Config["SHARE_CPU"].(bool)
	c.cond.L.Lock()
	if c.currentSlot < totalSlot {
		c.cond.L.Unlock()
		return false, nil, nil, nil, "", 0
	}
	sort.Slice(c.nodes, func(i int, j int) bool {
		return c.nodes[i].slot > c.nodes[j].slot
	})
	ip := make([]string, 0)
	port := make([]int, 0)
	cpuSet := make([]string, 0)
	totalSlotRet := totalSlot
	for _, node := range c.nodes {
		useSlot := node.slot
		if totalSlotRet < useSlot {
			useSlot = totalSlotRet
		}
		node.slot -= useSlot
		totalSlotRet -= useSlot
		var cpuArr []string
		var cpuStrBuilder strings.Builder
		var cpuP int = 0
		for i := 0; i < useSlot; i += slotPerContainer {
			ip = append(ip, node.ip)
			port = append(port, node.ports[0])
			node.ports = node.ports[1:]
			var cpu string
			for j := 0; j < slotPerContainer; j++ {
				for cpuP < len(node.cpuMask) {
					if !node.cpuMask[cpuP] {
						node.cpuMask[cpuP] = true
						cpu += strconv.Itoa(cpuP) + ","
						break
					}
					cpuP++
				}
			}
			cpuArr = append(cpuArr, cpu[:len(cpu)-1])
			cpuStrBuilder.WriteString(cpu)
		}
		cpuStr := cpuStrBuilder.String()
		cpuStr = cpuStr[:len(cpuStr)-1]
		for i := 0; i < useSlot; i += slotPerContainer {
			if shareCPU {
				cpuSet = append(cpuSet, cpuStr)
			} else {
				cpuSet = append(cpuSet, cpuArr[i/slotPerContainer])
			}
		}
		if totalSlotRet == 0 {
			break
		}
	}
	c.currentSlot -= totalSlot
	coordinatorIp := ip[0]
	coordinatorPort := c.nodes[0].ports[0]
	c.nodes[0].ports = c.nodes[0].ports[1:]
	c.cond.L.Unlock()
	return true, ip, port, cpuSet, coordinatorIp, coordinatorPort
}

func (c *Cluster) Put(slotPerContainer int, ip []string, port []int, cpuSet []string, coordinatorIp string, coordinatorPort int) {
	c.cond.L.Lock()
	ipPort := make(map[string][]int)
	ipCpuSet := make(map[string]string)
	for i := range ip {
		if _, ok := ipPort[ip[i]]; !ok {
			ipPort[ip[i]] = make([]int, 0)
		}
		ipPort[ip[i]] = append(ipPort[ip[i]], port[i])
		ipCpuSet[ip[i]] = cpuSet[i]
	}
	for _, node := range c.nodes {
		if ip, ok := ipPort[node.ip]; ok {
			node.ports = append(node.ports, ip...)
			node.slot += len(ipPort[node.ip]) * slotPerContainer
			c.currentSlot += len(ipPort[node.ip]) * slotPerContainer
		}
		if cpuSet, ok := ipCpuSet[node.ip]; ok {
			cpuList := strings.Split(cpuSet, ",")
			for _, cpu := range cpuList {
				idx, _ := strconv.Atoi(cpu)
				node.cpuMask[idx] = false
			}
		}
		if node.ip == coordinatorIp {
			node.ports = append(node.ports, coordinatorPort)
		}
	}
	c.cond.L.Unlock()
}
