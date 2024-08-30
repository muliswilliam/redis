package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

const PORT = 6379

func main() {
	addr := fmt.Sprintf("0.0.0.0:%d", PORT)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer l.Close()

	// create AOF
	aof, err := NewAof("database.aof")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer aof.Close()

	// read AOF and replace commands
	err = aof.Read(func(value Value) {
		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			return
		}
		handler(args)
	})
	if err != nil {
		log.Println("Error reading AOF: ", err)
	}

	fmt.Println("Listening for connections on:", addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go handleConnection(conn, aof)
	}
}

// handle each client connection
func handleConnection(conn net.Conn, aof *Aof) {
	defer conn.Close()

	for {
		resp := NewResp(conn)
		value, err := resp.Read()
		if err != nil {
			if err == io.EOF {
				// Client has disconnected
				fmt.Println("Client disconnected")
			} else {
				// Other errors
				fmt.Println("Error reading from client:", err)
			}
			return // Exit the goroutine
		}

		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]
		writer := NewWriter(conn)

		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			err := writer.Write(Value{typ: "string", str: ""})
			if err != nil {
				log.Println("Error writing value: ", err)
			}
			continue
		}

		if command == "SET" || command == "HSET" {
			err := aof.Write(value)
			if err != nil {
				log.Println("Error writing to AOF: ", err)
			}
		}

		result := handler(args)
		err = writer.Write(result)
		if err != nil {
			log.Println("Error writing value: ", err)
		}
	}
}
