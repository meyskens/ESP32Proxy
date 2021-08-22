package endpoints

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type EndpointDialer struct {
	socket *websocket.Conn
	mutex  sync.Mutex
}

func NewEndpointDialer(socket *websocket.Conn) *EndpointDialer {
	return &EndpointDialer{
		socket: socket,
	}
}

func (d *EndpointDialer) Request(req *http.Request) *http.Response {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.socket == nil {
		return &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       http.NoBody,
		}
	}

	// write request
	var b bytes.Buffer
	req.WriteProxy(&b)

	err := d.socket.WriteMessage(websocket.BinaryMessage, b.Bytes())
	if err != nil {
		log.Println(err)
		return &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       http.NoBody,
		}
	}

	_, message, err := d.socket.ReadMessage()
	if err != nil {
		log.Println(err)
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       http.NoBody,
		}
	}

	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewBuffer(message)), req)
	if err != nil {
		log.Println(err)
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       http.NoBody,
		}
	}

	return resp
}
