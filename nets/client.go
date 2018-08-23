package nets

import (
	"net"
	"bufio"
)

type Client struct {
	id int

	connection net.Conn // TCP连接

	udpConn       *net.UDPConn
	udpRemoteAddr *net.UDPAddr
	udpLocalAddr  *net.UDPAddr
}

func (client *Client) Id() int {
	return client.id
}

func (client *Client) WriteString(message string) {
	client.Write([]byte(message))
}

func (client *Client) Writeln(message string) {
	client.Write([]byte(message + "\n"))
}

func (client *Client) Write(bytes []byte) {
	if client.connection != nil {
		client.connection.Write(bytes)
	}
	if client.udpConn != nil && client.udpLocalAddr != nil {
		client.udpConn.WriteToUDP(bytes, client.udpLocalAddr)
	}
}

func (client *Client) Close() {
	if client.connection != nil {
		client.connection.Close()
	}
}

func (client *Client) Connect(network string, address string) error {
	conn, err := net.Dial(network, address)
	if err != nil {
		return err
	}
	client.connection = conn
	return nil
}

func (client *Client) Receive(receiver func(data []byte)) {
	scanner := bufio.NewScanner(client.connection)
	for scanner.Scan() {
		receiver(scanner.Bytes())
	}
}

func (client *Client) RemoteAddr() net.Addr {
	if client.udpRemoteAddr != nil {
		return client.udpRemoteAddr
	}
	return client.connection.RemoteAddr()
}

func (client *Client) LocalAddr() net.Addr {
	if client.udpLocalAddr != nil {
		return client.udpLocalAddr
	}
	return client.connection.LocalAddr()
}
