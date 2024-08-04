package main

import (
	"fmt"
	"net"
)

const PORT = 6379

func main() {
	addr := fmt.Sprintf("0.0.0.0:%d", PORT)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Listening for connections on:", addr)

	conn, err := l.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()

	for {
		resp := NewResp(conn)
		value, err := resp.Read()
		if err != nil {
			fmt.Println("Error reading fro client: ", err)
			return
		}
		fmt.Println(value)
		// write
		conn.Write([]byte("+PONG\r\n"))
	}
}
