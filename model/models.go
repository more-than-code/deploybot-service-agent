package model

type BuildConfig struct {
	ImageName  string             `json:"imageName"`
	ImageTag   string             `json:"imageTag" bson:",omitempty"`
	Args       map[string]*string `json:"args" bson:",omitempty"`
	Dockerfile string             `json:"dockerfile" bson:",omitempty"`
	RepoUrl    string             `json:"repoUrl"`
	RepoName   string             `json:"repoName"`
	RepoBranch string             `json:"repoBranch"`
}

type RestartPolicy struct {
	Name              string `json:"name" bson:",omitempty"`
	MaximumRetryCount int    `json:"maxiumRetryCount" bson:",omitempty"`
}

type DeployConfig struct {
	ImageName     string            `json:"imageName"`
	ImageTag      string            `json:"imageTag" bson:",omitempty"`
	ServiceName   string            `json:"serviceName" bson:",omitempty"`
	VolumeMounts  map[string]string `json:"volumeMounts" bson:",omitempty"`
	Files         map[string]string `json:"files" bson:",omitempty"`
	AutoRemove    bool              `json:"autoRemove"`
	RestartPolicy RestartPolicy     `json:"restartPolicy" bson:",omitempty"`
	Env           []string          `json:"env" bson:",omitempty"`
	Ports         map[string]string `json:"ports" bson:",omitempty"`
	NetworkId     string            `json:"networkId" bson:",omitempty"`
	NetworkName   string            `json:"networkName" bson:",omitempty"`
	Command       string            `json:"command" bson:",omitempty"`
	Links         []string          `json:"links" bson:",omitempty"`
}

type Network struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type GetNetworkResponse struct {
	Code    int      `json:"code"`
	Msg     string   `json:"msg"`
	Payload *Network `json:"payload"`
}

type GetNetworksResponse struct {
	Code    int       `json:"code"`
	Msg     string    `json:"msg"`
	Payload []Network `json:"payload"`
}

type CreateNetworkResponse struct {
	Code    int      `json:"code"`
	Msg     string   `json:"msg"`
	Payload *Network `json:"payload"`
}

type CreateNetworkInput struct {
	Name string `json:"name"`
}

type DeleteNetworkResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type GetDiskInfoResponse struct {
	Msg     string    `json:"msg"`
	Code    int       `json:"code"`
	Payload *DiskInfo `json:"payload"`
}

type DiskInfo struct {
	TotalSize uint64 `json:"totalSize"`
	AvailSize uint64 `json:"availSize"`
	Path      string `json:"path"`
}

type RestartConfig struct {
	ServiceName string `json:"serviceName"`
}
