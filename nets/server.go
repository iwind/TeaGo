package nets

import (
	"net"
	"log"
	"bufio"
	"github.com/iwind/TeaGo/logs"
)

type Server struct {
	network string
	address string

	listener net.Listener
	udpConn  *net.UDPConn

	onAccept  func(client *Client)
	onClose   func(client *Client)
	onReceive func(client *Client, data []byte)
}

func NewServer(network, address string) *Server {
	return &Server{
		network: network,
		address: address,
	}
}

func (server *Server) AcceptClient(callback func(client *Client)) {
	server.onAccept = callback
}

func (server *Server) CloseClient(callback func(client *Client)) {
	server.onClose = callback
}

func (server *Server) ReceiveClient(callback func(client *Client, data []byte)) {
	server.onReceive = callback
}

func (server *Server) Listen() error {
	switch server.network {
	case "udp", "udp4", "udp6":
		return server.listenUDP()
	default:
		return server.listenTCP()
	}

	return nil
}

func (server *Server) listenTCP() error {
	listener, err := net.Listen(server.network, server.address)
	if err != nil {
		return err
	}
	server.listener = listener

	var id int
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		id ++

		var client = &Client{
			id:         id,
			connection: conn,
		}
		if server.onAccept != nil {
			server.onAccept(client)
		}
		go func(client *Client) {
			input := bufio.NewScanner(client.connection)
			for input.Scan() {
				if server.onReceive != nil {
					server.onReceive(client, input.Bytes())
				}
			}

			defer func() {
				client.connection.Close()
				if server.onClose != nil {
					server.onClose(client)
				}
			}()
		}(client)
	}
}

func (server *Server) listenUDP() error {
	addr, err := net.ResolveUDPAddr(server.network, server.address)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP(server.network, addr)
	if err != nil {
		return err
	}
	server.udpConn = conn

	for {
		data := make([]byte, 1024)
		n, remoteAddr, err := conn.ReadFromUDP(data)
		if err != nil {
			logs.Error(err)
			continue
		}

		client := &Client{
			udpConn:       conn,
			udpRemoteAddr: addr,
			udpLocalAddr:  remoteAddr,
		}
		if server.onReceive != nil {
			server.onReceive(client, data[:n])
		}
	}
}

func (server *Server) Close() error {
	if server.udpConn != nil {
		return server.udpConn.Close()
	}
	if server.listener != nil {
		return server.listener.Close()
	}
	return nil
}
