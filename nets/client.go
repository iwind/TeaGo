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

func (this *Client) Id() int {
	return this.id
}

func (this *Client) WriteString(message string) {
	this.Write([]byte(message))
}

func (this *Client) Writeln(message string) {
	this.Write([]byte(message + "\n"))
}

func (this *Client) Write(bytes []byte) {
	if this.connection != nil {
		this.connection.Write(bytes)
	}
	if this.udpConn != nil && this.udpLocalAddr != nil {
		this.udpConn.WriteToUDP(bytes, this.udpLocalAddr)
	}
}

func (this *Client) Close() {
	if this.connection != nil {
		this.connection.Close()
	}
}

func (this *Client) Connect(network string, address string) error {
	conn, err := net.Dial(network, address)
	if err != nil {
		return err
	}
	this.connection = conn
	return nil
}

func (this *Client) Receive(receiver func(data []byte)) {
	scanner := bufio.NewScanner(this.connection)
	for scanner.Scan() {
		receiver(scanner.Bytes())
	}
}

func (this *Client) RemoteAddr() net.Addr {
	if this.udpRemoteAddr != nil {
		return this.udpRemoteAddr
	}
	return this.connection.RemoteAddr()
}

func (this *Client) LocalAddr() net.Addr {
	if this.udpLocalAddr != nil {
		return this.udpLocalAddr
	}
	return this.connection.LocalAddr()
}
