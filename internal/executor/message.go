package executor

// Messages for monitor-write widgets can be found in the cmd/monitor_write and
// internal/fileio packages.

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
