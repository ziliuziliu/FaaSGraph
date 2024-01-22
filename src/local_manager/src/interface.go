package main

type CreateContainerRequest struct {
	ContainerType string `json:"container_type"`
	App           string `json:"app"`
	Function      string `json:"function"`
	Port          int32  `json:"port"`
	CpuSet        string `json:"cpu_set"`
	DataPath      string `json:"data_path"`
}

type RemoveContainerRequest struct {
	Id string `json:"id"`
}

type ResetCPUShareRequest struct {
	Id       string `json:"id"`
	CPUShare int    `json:"cpu_share"`
}
