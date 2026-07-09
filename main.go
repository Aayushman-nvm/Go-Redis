package main

import (
	"fmt"
	"net"
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
		fmt.Println(value)

		//Once read successfully, we send back an "OK" message to the client
		writer := NewWriter(conn)
		writer.Write(Value{typ: "string", str: "OK"})
	}

}
