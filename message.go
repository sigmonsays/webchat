package webchat

import (
	"encoding/json"
)

type OpCode int

const (
	MessageOp OpCode = iota
	HistoryOp
	NoticeOp
	JoinOp
	NickOp
)

type Message struct {
	connection *connection
	Id         int64  `json:"id"`
	Op         OpCode `json:"op"`
	From       string `json:"from"`
	Message    string `json:"message"`

	// pop up a notification
	Notify bool `json:"notify"`
}

func (m *Message) Json() []byte {
	data, _ := json.Marshal(m)
	return data
}
func (m *Message) FromJson(data []byte) error {
	err := json.Unmarshal(data, m)
	return err
}
