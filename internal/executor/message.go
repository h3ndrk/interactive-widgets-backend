package executor

type MonitorWriteOutputMessage struct {
	Contents []byte   `json:"contents"`
	Errors   [][]byte `json:"errors"`
}

type MonitorWriteInputMessage struct {
	Contents []byte `json:"contents"`
}

type ButtonClickMessage struct {
	Click bool `json:"click"`
}

type OutputStream string

const StdoutStream OutputStream = "stdout"
const StderrStream OutputStream = "stderr"

type ButtonOutputMessage struct {
	Origin OutputStream `json:"origin"`
	Data   []byte       `json:"data"`
}

type TerminalInputMessage struct {
	Data []byte `json:"data"`
}

type TerminalOutputMessage struct {
	Data []byte `json:"data"`
}
