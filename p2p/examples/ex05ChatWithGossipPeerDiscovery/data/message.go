package data

import (
	"fmt"
)

// Message ..
type Message struct {
	Message string
	ID      int
	Origin  string
}

// WriteData for debug
func (m *Message) WriteData(pre string, message *Message) {
	fmt.Println("---------------------")
	fmt.Printf("*** %s ***\n", pre)
	fmt.Printf("ID:%d\n Origin:%s\n Message:%s", message.ID, message.Origin, message.Message)
	fmt.Println("---------------------")
}

// Key ..
func (m *Message) Key() string {
	return fmt.Sprintf("%s|%d", m.Origin, m.ID)
}

// CreateMessage .
func CreateMessage(data string, address string, id int) *Message {
	message := new(Message)
	message.Message = data
	message.ID = id
	message.Origin = address
	return message
}

// Store ...
type Store map[string]*Message
