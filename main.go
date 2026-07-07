package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	//Creating a tcp listener to accept tcp connection request
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}

	//Waiting for a client to connect to our listener
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()

	for {
		//Our buffer
		buf := make([]byte, 1024)

		//From client connection, we read <buf> bytes at a time
		_, err := conn.Read(buf)
		if err != nil {
			//If we reach EOF (End of file)... good, no problem
			if err == io.EOF {
				break
			}
			fmt.Println("error reading from client: ", err.Error())
			os.Exit(1)
		}

		//Once read successfully, we send back an "OK" message to the client
		conn.Write([]byte("+OK\r\n"))
	}

	fmt.Println("Hello World")
}
