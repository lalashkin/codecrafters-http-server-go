package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	
	for {
		conn, err := l.Accept()

		go handleConnection(conn)

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
	}
}

func handleConnection(conn net.Conn) {

	defer conn.Close()

	var recv []byte

	_, err := conn.Read(recv)
	if err != nil {
		panic(err)
	}

	statOk := "HTTP/1.1 200 OK\r\n\r\n"

	_, err = conn.Write([]byte(statOk))
	if err != nil {
		panic(err)
	}

}