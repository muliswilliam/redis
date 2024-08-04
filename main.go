package main

import (
	"fmt"
	"io"
	"net"
	"os"
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
		buf := make([]byte, 1024)

		// read from client
		_, err = conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading from client: ", err.Error())
			os.Exit(1)
		}

		conn.Write([]byte("+PONG\r\n"))
	}
}
