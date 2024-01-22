package common

import (
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
)

type FunctionInfo struct {
	App                 string
	Function            string
	MaxSlotPerContainer int
}

var Config map[string]any = make(map[string]any)
var Info map[string]FunctionInfo = make(map[string]FunctionInfo)
var WorkerList []string = make([]string, 0)

func LoadConfig() {
	data, err := ioutil.ReadFile("/home/ubuntu/FaaSGraph/config/config.json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, &Config)
	appList := Config["APP"].([]any)
	for _, appRaw := range appList {
		app := appRaw.(map[string]any)
		appName := app["NAME"].(string)
		for _, functionRaw := range app["FUNCTION"].([]any) {
			function := functionRaw.(map[string]any)
			functionName := function["NAME"].(string)
			maxSlotPerContainer := 1
			if v, ok := function["MAX_SLOT_PER_CONTAINER"]; ok {
				maxSlotPerContainer = int(v.(float64))
			}
			Info[appName+"-"+functionName] = FunctionInfo{
				App:                 appName,
				Function:            functionName,
				MaxSlotPerContainer: maxSlotPerContainer,
			}
		}
	}
	for _, ip := range Config["WORKER"].([]any) {
		WorkerList = append(WorkerList, ip.(string))
	}
}

func LoadGraphConfig(graph string) map[string]any {
	graphConfig := map[string]any(Config["GRAPH"].(map[string]any)[graph].(map[string]any))
	partition := int32(graphConfig["PARTITION"].(float64))
	startNodesF, _ := os.Open(graphConfig["DATA_PATH"].(string) + "/start_nodes_" + strconv.Itoa(int(graphConfig["PARTITION"].(float64))) + ".bin")
	startNodes := make([]int32, 0)
	sizes := make(map[string]int32)
	b := make([]byte, 4)
	for i := int32(0); i < partition; i++ {
		startNodesF.Read(b)
		startNode := int32(binary.LittleEndian.Uint32(b))
		startNodes = append(startNodes, startNode)
	}
	for i := int32(0); i < partition; i++ {
		var size int32 = 0
		startNode := startNodes[i]
		key := strconv.Itoa(int(startNode))
		if i == partition-1 {
			size = int32(graphConfig["TOTAL_NODE"].(float64)) - startNodes[i]
		} else {
			size = startNodes[i+1] - startNodes[i]
		}
		sizes[key] = size
	}
	newGraphConfig := make(map[string]any)
	for k, v := range graphConfig {
		newGraphConfig[k] = v
	}
	newGraphConfig["START_NODES"] = startNodes
	newGraphConfig["SIZES"] = sizes
	return newGraphConfig
}

func init() {
	LoadConfig()
}
