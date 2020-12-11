package main

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	for {
		conn, err := openWebSocket("127.0.0.1:8041")
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		listeningOnWebSocket(conn)
	}
}

const (
	ws = "ws"
)

func openWebSocket(address string) (*websocket.Conn, error) {
	scheme := ws

	u := url.URL{
		Scheme: scheme,
		Host:   address,
		Path:   "/indexer",
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func listeningOnWebSocket(conn *websocket.Conn) {
	for {
		msgType, message, err := conn.ReadMessage()
		if msgType == websocket.CloseMessage {
			return
		}
		if err == nil {
			fmt.Println("########## Message received ##########")
			fmt.Println(string(message))
			continue
		}

		_, _ = err.(*websocket.CloseError)
		return
	}
}
