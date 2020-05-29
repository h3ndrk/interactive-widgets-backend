package executor

type MonitorWriteOutputMessage struct {
	Contents []byte   `json:"contents"`
	Errors   [][]byte `json:"errors"`
}

type MonitorWriteInputMessage struct {
	Contents []byte `json:"contents"`
}
