package main

import (
	"fmt"
	"io"
	"net"
	"os"
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
		fmt.Printf("Listening on port :%d", PORT)
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
		//Our buffer
		buf := make([]byte, 1024)

		//From client connection, we read <buf> bytes at a time
		//_, err := conn.Read(buf)
		n, err := conn.Read(buf)
		if err != nil {
			//If we reach EOF (End of file)... good, no problem
			if err == io.EOF {
				break
			}
			fmt.Println("error reading from client: ", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Received %q\n", buf[:n])

		//Once read successfully, we send back an "OK" message to the client
		conn.Write([]byte("+OK\r\n"))
		fmt.Println("Sent OK")
	}

	fmt.Println("Hello World")
}
