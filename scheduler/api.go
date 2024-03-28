package scheduler

type Network struct {
	Name string
	Id   string
}

type GetNetworkResponse struct {
	Code    int      `json:"code"`
	Msg     string   `json:"msg"`
	Payload *Network `json:"payload"`
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
