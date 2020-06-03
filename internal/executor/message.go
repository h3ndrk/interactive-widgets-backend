package executor

// MonitorWriteContentsMessage gets send from a monitor-write widget to the
// client.
type MonitorWriteContentsMessage struct {
	Contents []byte `json:"contents"`
}

// MonitorWriteErrorMessage gets send from a monitor-write widget to the
// client.
type MonitorWriteErrorMessage struct {
	Error []byte `json:"error"`
}

// MonitorWriteInputMessage gets send from the client to a monitor-write
// widget.
type MonitorWriteInputMessage struct {
	Contents []byte `json:"contents"`
}

// ButtonClickMessage gets send from the client to a button widget.
type ButtonClickMessage struct {
	Click bool `json:"click"`
}

// OutputStream represents the output stream.
type OutputStream string

// StdoutStream represents the stdout stream.
const StdoutStream OutputStream = "stdout"

// StderrStream represents the stderr stream.
const StderrStream OutputStream = "stderr"

// ButtonOutputMessage gets send from a button widget to the client.
type ButtonOutputMessage struct {
	Origin OutputStream `json:"origin"`
	Data   []byte       `json:"data"`
}

// ButtonClearMessage gets send from a button widget to the client.
type ButtonClearMessage struct {
	Clear bool `json:"clear"`
}

// TerminalInputMessage gets send from the client to a terminal widget.
type TerminalInputMessage struct {
	Data []byte `json:"data"`
}

// TerminalOutputMessage gets send from a terminal widget to the client.
type TerminalOutputMessage struct {
	Data []byte `json:"data"`
}
