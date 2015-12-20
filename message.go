//go:generate stringer -type=OpCode

package webchat

import (
	"encoding/json"
)

type OpCode int

const (
	InvalidOp OpCode = iota

	// client has connected
	RegisterOp

	// client has disconnected
	UnregisterOp

	// a message has been sent
	MessageOp

	// a notice is a informational message. likely from the system
	NoticeOp
	// a user has joined
	JoinOp

	// a user has changed their nick name
	NickOp
)

type Message struct {
	connection *Connection
	Id         int64  `json:"id"`
	Op         OpCode `json:"op"`
	From       string `json:"from"`
	Message    string `json:"message"`
}

func (m *Message) Json() []byte {
	data, _ := json.Marshal(m)
	return data
}
func (m *Message) FromJson(data []byte) error {
	err := json.Unmarshal(data, m)
	return err
}
