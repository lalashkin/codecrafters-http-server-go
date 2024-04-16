package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

// HTTPReq is arriving HTTP request
type HTTPReq struct {
	method  string
	path    string
	ver     string
	headers []string
	body    string
}

// HTTPResp is outbound HTTP response
type HTTPResp struct {
	version string
	status string
	headers []string
	body string
}

func (resp HTTPResp) String() string {

	var s strings.Builder

	s.WriteString(fmt.Sprintf("%s %s\r\n", resp.version, resp.status))

	for _, header := range resp.headers {
		s.WriteString(fmt.Sprintf("%s\r\n", header))
	}

	s.WriteString("\r\n")
	s.WriteString(fmt.Sprintf("%s", resp.body))

	return s.String()
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

	req := readPayload(conn)

	resp := HTTPResp{
		version: "HTTP/1.1",
		headers: []string{
			"Server: lalashkin/0.0.1",
		},
	}

	switch  {
	case req.path == "/":
		resp.status = "200 OK"
		resp.headers = append(resp.headers,
			"Content-Length: 0",
		)

	case strings.HasPrefix(req.path, "/echo"):

		reqContent := strings.Split(req.path, "/echo/")[1]

		resp.status = "200 OK"
		resp.headers = append(resp.headers,
			"Content-Type: text/plain",
			fmt.Sprintf("Content-Length: %d", len(reqContent)),
		)
		resp.body = reqContent

	default:
		resp.status = "404 Not Found"
		resp.headers = append(resp.headers,
			"Content-Length: 0",
		)
	}

	log.Printf("Composed response obj: %s", resp)

	_, err := conn.Write([]byte(resp.String()))
	if err != nil {
		panic(err)
	}

}

func readPayload(conn net.Conn) HTTPReq {

	reader := bufio.NewReader(conn)

	bytesRead := 0
	index := 0
	var req HTTPReq

	for {
		line, err := reader.ReadBytes('\n')

		if err != nil {
			log.Printf("Reader stopped: %s", err)
			break
		}

		n := len(line)
		bytesRead += n
		
		if bytes.HasSuffix(line, []byte("\r\n")) {

			trLine := line[:n - 2]
			if len(trLine) == 0 {
				break
			}

			switch index {
			case 0:
				startLine := bytes.Split(trLine, []byte(" "))
				req = HTTPReq{
					method: string(startLine[0]),
					path:   string(startLine[1]),
					ver:    string(startLine[2]),
				}
			default:
				req.headers = append(req.headers, string(trLine))
			}
		}

		log.Printf("Line: %s", line)
		index++
	}

	log.Printf("Composed request obj: %s", req)
	return req
}
