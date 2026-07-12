package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

//test command:- redis-cli -p 6380.
// This keeps my redis service and my clone seperate while letting the cli client comunicate on respective ports,
// helping me in comparing both side by side while building

func main() {
	//Creating a tcp listener to accept tcp connection request
	PORT := 6380
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Printf("Listening on port :%d\n", PORT)
	}

	aof, err := NewAof("database.aof")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer aof.Close()

	aof.Read(func(value Value) {
		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			return
		}

		handler(args)
	})

	fmt.Println("SETs:", SETs)
	fmt.Println("HSETs:", HSETs)

	fmt.Println("Waiting for connection...")
	//Waiting for a client to connect to our listener
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("Client connected:", conn.RemoteAddr())
	}

	defer conn.Close()

	for {
		resp := NewResp(conn)

		value, err := resp.Read()
		if err != nil {
			fmt.Println(err)
			return
		}

		if value.typ != "array" {
			fmt.Println("Invalid request, expected array")
			continue
		}

		if len(value.array) == 0 {
			fmt.Println("Invalid request, expected array length > 0")
			continue
		}

		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		fmt.Println(value)

		//Once read successfully, we send back an "OK" message to the client
		writer := NewWriter(conn)

		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			writer.Write(Value{typ: "string", str: ""})
			continue
		}

		if command == "SET" || command == "HSET" {
			if err := aof.Write(value); err != nil {
				log.Println(err)
			}
		}

		result := handler(args)
		writer.Write(result)
	}

}
