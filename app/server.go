package main

import (
	"bufio"
	"net"
	"os"
	"log"
	"strings"
)

// Req is arriving HTTP request
type Req struct {
	method string
	path string
	ver string
}

func main() {
	log.Println("Starting ad-hoc TCP server...")
	
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		log.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	
	for {
		conn, err := l.Accept()

		go handleConnection(conn)

		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reqString := readPayload(conn)

	reqStringSlice := strings.Split(reqString, "\r\n")

	startLineSlice := strings.Split(reqStringSlice[0], " ")

	req := &Req{method: startLineSlice[0], path: startLineSlice[1], ver: startLineSlice[2]}

	log.Printf("Composed request obj: %s", req)

	var resp string

	switch req.path {
	case "/":
		resp = "HTTP/1.1 200 OK\r\n\r\n"
	case "/index.html":
		resp = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	_, err := conn.Write([]byte(resp))
	if err != nil {
		panic(err)
	}

}

func readPayload(conn net.Conn) string {

	reader := bufio.NewReader(conn)
	recv := make([]byte, 1024)

	_, err := reader.Read(recv)

	if err != nil {
		panic(err)
	}

	log.Printf("Read: %s", string(recv))
	return string(recv)
}