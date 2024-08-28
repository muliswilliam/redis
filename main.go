package main

import (
	"fmt"
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

	// create AOFju
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
			fmt.Println("Error reading from client: ", err)
			return
		}

		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]
		writer := NewWriter(conn)

		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			err := writer.Write(Value{typ: "string", str: ""})
			if err != nil {
				log.Println("Errr writing value: ", err)
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
			log.Println("Errr writing value: ", err)
		}
	}
}
