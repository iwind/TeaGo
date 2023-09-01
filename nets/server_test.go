package nets

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	var server = NewServer("tcp", "localhost:8001")
	var clients = map[int]*Client{}
	var mux sync.Mutex
	server.AcceptClient(func(client *Client) {
		mux.Lock()
		clients[client.id] = client
		mux.Unlock()
		log.Println("accept ", client.id)
	})
	server.ReceiveClient(func(client *Client, data []byte) {
		message := string(data)
		if strings.TrimSpace(message) == "quit" {
			client.Close()
			return
		}
		client.Writeln("OK")
		log.Println("receive ", client.id, " ", message)
		for i, c := range clients {
			if i != client.id {
				go func(i int, c *Client) {
					log.Println("write message to ", i)
					now := time.Now()
					c.WriteString(message + "|" + fmt.Sprintf("%d", now.Nanosecond()) + "\n")
				}(i, c)
			}
		}
	})
	server.CloseClient(func(client *Client) {
		mux.Lock()
		delete(clients, client.id)
		log.Println("close ", client.id, " left:", len(clients))
		mux.Unlock()
	})
	server.Listen()
}

func TestNewServerSimple(t *testing.T) {
	var server = NewServer("tcp", "localhost:8001")
	var clients = map[int]*Client{}
	var mux sync.Mutex
	server.AcceptClient(func(client *Client) {
		mux.Lock()
		clients[client.id] = client
		mux.Unlock()
		log.Println("clients:", len(clients))
	})
	server.ReceiveClient(func(client *Client, data []byte) {
		message := string(data)
		log.Println(message)
		client.Writeln("OK")
	})
	server.CloseClient(func(client *Client) {
		mux.Lock()
		delete(clients, client.id)
		mux.Unlock()
	})
	server.Listen()
}

func TestClient(t *testing.T) {
	client := &Client{}
	err := client.Connect("tcp", "localhost:8001")
	if err != nil {
		t.Fatal(err)
	}
	go client.Writeln("Hello")
	client.Receive(func(data []byte) {
		message := string(data)
		log.Println("received:", message)
	})
	t.Log("finished")
}

func TestMultipleClients(t *testing.T) {
	for i := 0; i < 2000; i ++ {
		go func(i int) {
			client := &Client{}
			err := client.Connect("tcp", "localhost:8001")
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("connected:", i)
			client.Receive(func(data []byte) {
				message := string(data)
				log.Println("received:", message)
			})
			t.Log("finished ", i)
		}(i)
	}
	time.Sleep(20 * time.Second)
}
