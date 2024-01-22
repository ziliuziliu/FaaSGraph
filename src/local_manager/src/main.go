package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var dockerCli *client.Client
var Config map[string]any = make(map[string]any)
var pool map[string]*MemPool = make(map[string]*MemPool)
var poolLock sync.Mutex
var ctx = context.Background()
var overallComm int64 = 0

func OKResult() []byte {
	result := map[string]string{"status": "OK"}
	rep, _ := json.Marshal(result)
	return rep
}

func init() {
	loadConfig()
	resetContainer()
}

func loadConfig() {
	data, err := ioutil.ReadFile("/home/ubuntu/FaaSGraph/config/config.json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, &Config)
}

func resetContainer() {
	fmt.Println("clearing previous container")
	cmd := exec.Command("/bin/sh", "-c", "docker rm -f $(docker ps -aq --filter label=label=graph)")
	output, _ := cmd.Output()
	fmt.Println(string(output))
	dockerCli, _ = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

func createContainer(writer http.ResponseWriter, request *http.Request) {
	req, _ := ioutil.ReadAll(request.Body)
	r := CreateContainerRequest{}
	err := json.Unmarshal(req, &r)
	if err != nil {
		panic(err)
	}
	fmt.Println("create container for", r.ContainerType, r.App, r.Function, "at", r.Port, "cpu", r.CpuSet)

	// portStr := make([]string, 0)
	// for i := 0; i < 12; i++ {
	// 	portStr = append(portStr, "0.0.0.0:"+strconv.Itoa(int(r.Port)+i)+":"+strconv.Itoa(5000+i))
	// }
	// _, portMapping, err := nat.ParsePortSpecs(portStr)

	_, portMapping, err := nat.ParsePortSpecs([]string{"0.0.0.0:" + strconv.Itoa(int(r.Port)) + ":5000"})
	labels := map[string]string{"label": "graph"}
	if err != nil {
		panic(err)
	}
	rep := make(map[string]string)
	pathBinding := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: r.DataPath,
			Target: "/lambda_executor/data",
		},
	}
	resources := container.Resources{
		CpusetCpus: r.CpuSet,
	}
	imgName := r.App + "-coordinator"
	if r.ContainerType == "graph" {
		imgName = r.App + "-" + r.Function
	}
	for {
		resp, err := dockerCli.ContainerCreate(ctx, &container.Config{
			Image:  imgName,
			Labels: labels,
		}, &container.HostConfig{
			PortBindings: portMapping,
			Mounts:       pathBinding,
			IpcMode:      "host",
			Resources:    resources,
		}, nil, nil, "")
		if err == nil {
			err := dockerCli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
			if err == nil {
				rep["id"] = resp.ID
				break
			} else {
				fmt.Println(err)
			}
		} else {
			fmt.Println(err)
		}
	}
	repBody, err := json.Marshal(rep)
	if err != nil {
		panic(err)
	}
	writer.Write(repBody)
}

func removeContainer(writer http.ResponseWriter, request *http.Request) {
	req, _ := ioutil.ReadAll(request.Body)
	r := RemoveContainerRequest{}
	err := json.Unmarshal(req, &r)
	if err != nil {
		panic(err)
	}
	dockerCli.ContainerRemove(context.Background(), r.Id, types.ContainerRemoveOptions{
		Force: true,
	})
	writer.Write(OKResult())
}

func allocateMem(writer http.ResponseWriter, request *http.Request) {
	requestId := request.Header.Get("Request-Id")
	totalNode, _ := strconv.Atoi(request.Header.Get("Total-Node"))
	vertexArrayCnt, _ := strconv.Atoi(request.Header.Get("Vertex-Array-Cnt"))
	poolLock.Lock()
	pool[requestId] = NewMemPool(requestId, totalNode, vertexArrayCnt)
	poolLock.Unlock()
	writer.Write(OKResult())
}

func clearMem(writer http.ResponseWriter, request *http.Request) {
	requestId := request.Header.Get("Request-Id")
	poolLock.Lock()
	pool[requestId].ClearMem()
	delete(pool, requestId)
	poolLock.Unlock()
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(overallComm))
	writer.Write(b)
}

func setVertexInfo(writer http.ResponseWriter, request *http.Request) {
	comm := uint32(0)
	b := make([]byte, 4)
	requestId := request.Header.Get("Request-Id")
	aggregateOp := request.Header.Get("Aggregate-Op")
	tag, _ := strconv.Atoi(request.Header.Get("Tag"))
	for {
		_, err := request.Body.Read(b)
		if err != nil {
			break
		}
		num := binary.LittleEndian.Uint32(b)
		comm += num
		data := make([]byte, num*8)
		for i := uint32(0); i < num; i++ {
			request.Body.Read(data[8*i : 8*i+4])
			request.Body.Read(data[8*i+4 : 8*i+8])
		}
		pool[requestId].AggregateVertexVal(tag, aggregateOp, ByteSlice2Uint32Slice(data))
	}
	writer.Write(OKResult())
	atomic.AddInt64(&overallComm, int64(comm))
}

func main() {
	http.HandleFunc("/create_container", createContainer)
	http.HandleFunc("/remove_container", removeContainer)
	http.HandleFunc("/allocate_mem", allocateMem)
	http.HandleFunc("/clear_mem", clearMem)
	http.HandleFunc("/set_vertex_info", setVertexInfo)
	fmt.Println("server started at 0.0.0.0:20000")
	err := http.ListenAndServe(":20000", nil)
	if err != nil {
		panic(err)
	}
}
