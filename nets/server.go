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

func (this *Server) AcceptClient(callback func(client *Client)) {
	this.onAccept = callback
}

func (this *Server) CloseClient(callback func(client *Client)) {
	this.onClose = callback
}

func (this *Server) ReceiveClient(callback func(client *Client, data []byte)) {
	this.onReceive = callback
}

func (this *Server) Listen() error {
	switch this.network {
	case "udp", "udp4", "udp6":
		return this.listenUDP()
	default:
		return this.listenTCP()
	}

	return nil
}

func (this *Server) listenTCP() error {
	listener, err := net.Listen(this.network, this.address)
	if err != nil {
		return err
	}
	this.listener = listener

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
		if this.onAccept != nil {
			this.onAccept(client)
		}
		go func(client *Client) {
			input := bufio.NewScanner(client.connection)
			for input.Scan() {
				if this.onReceive != nil {
					this.onReceive(client, input.Bytes())
				}
			}

			defer func() {
				client.connection.Close()
				if this.onClose != nil {
					this.onClose(client)
				}
			}()
		}(client)
	}
}

func (this *Server) listenUDP() error {
	addr, err := net.ResolveUDPAddr(this.network, this.address)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP(this.network, addr)
	if err != nil {
		return err
	}
	this.udpConn = conn

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
		if this.onReceive != nil {
			this.onReceive(client, data[:n])
		}
	}
}

func (this *Server) Close() error {
	if this.udpConn != nil {
		return this.udpConn.Close()
	}
	if this.listener != nil {
		return this.listener.Close()
	}
	return nil
}
