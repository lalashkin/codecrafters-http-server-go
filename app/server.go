package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// HTTPReq is arriving HTTP request
type HTTPReq struct {
	method  string
	path    string
	ver     string
	headers map[string]string
	body    string
}

// HTTPResp is outbound HTTP response
type HTTPResp struct {
	version string
	status  string
	headers map[string]string
	body    string
}

func (resp HTTPResp) String() string {

	var s strings.Builder

	s.WriteString(fmt.Sprintf("%s %s\r\n", resp.version, resp.status))

	for hName, hVal := range resp.headers {
		s.WriteString(fmt.Sprintf("%s: %s\r\n", hName, hVal))
	}

	s.WriteString("\r\n")
	s.WriteString(fmt.Sprintf("%s", resp.body))

	return s.String()
}

func main() {
	log.Println("Starting ad-hoc TCP server...")
	
	var servingDir string
	flag.StringVar(&servingDir, "directory", "", "Directory to serve files from")
	flag.Parse()

	if servingDir != "" {
		log.Printf("Serving files from %s", servingDir)
	}

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		log.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()

		go handleConnection(conn, servingDir)

		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
	}
}

func handleConnection(conn net.Conn, servingDir string) {
	defer conn.Close()

	req := readPayload(conn)

	resp := HTTPResp{
		version: "HTTP/1.1",
		headers: make(map[string]string),
	}

	switch {
	case req.path == "/":
		resp.status = "200 OK"
		resp.headers["Content-Length"] = "0"

	case strings.HasPrefix(req.path, "/files"):

		srvFile, _ := strings.CutPrefix(req.path, "/files/")
		srvFile = filepath.Join(servingDir, srvFile)

		if _, err := os.Stat(srvFile); err == nil {
			data, err := os.ReadFile(srvFile)
			if err != nil {
				log.Fatalf("Cannot open %s", srvFile)
			}

			resp.status = "200 OK"
			resp.headers["Content-Type"] = "application/octet-stream"
			resp.headers["Content-Length"] = fmt.Sprintf("%d", len(data))
			resp.body = string(data)

		} else if errors.Is(err, os.ErrNotExist) {
			resp.status = "404 Not Found"
			resp.headers["Content-Length"] = "0"
		} else {
			log.Fatalf("Error applying stat for %s", srvFile)
		}


	case strings.HasPrefix(req.path, "/echo"):

		reqContent, _ := strings.CutPrefix(req.path, "/echo/")

		resp.status = "200 OK"
		resp.headers["Content-Type"] = "text/plain"
		resp.headers["Content-Length"] = fmt.Sprintf("%d", len(reqContent))
		resp.body = reqContent

	case strings.HasPrefix(req.path, "/user-agent"):

		reqUAgent := req.headers["User-Agent"]

		resp.status = "200 OK"
		resp.headers["Content-Type"] = "text/plain"
		resp.headers["Content-Length"] = fmt.Sprintf("%d", len(reqUAgent))
		resp.body = reqUAgent

	default:
		resp.status = "404 Not Found"
		resp.headers["Content-Length"] = "0"
	}

	log.Printf("Response:\n%s", resp)

	_, err := conn.Write([]byte(resp.String()))
	if err != nil {
		panic(err)
	}

}

func readPayload(conn net.Conn) HTTPReq {

	reader := bufio.NewReader(conn)

	index := 0
	req := HTTPReq{
		headers: make(map[string]string),
	}

	for {
		line, err := reader.ReadBytes('\n')

		if err != nil {
			log.Printf("Reader stopped: %s", err)
			break
		}

		n := len(line)

		if bytes.HasSuffix(line, []byte("\r\n")) {

			trLine := line[:n-2]
			if len(trLine) == 0 {
				break
			}

			switch index {
			case 0:
				startLine := bytes.Split(trLine, []byte(" "))
				req.method = string(startLine[0])
				req.path = string(startLine[1])
				req.ver = string(startLine[2])
			default:
				hLine := string(trLine)
				hName := string(strings.Split(hLine, ": ")[0])
				hVal := string(strings.Split(hLine, ": ")[1])
				req.headers[hName] = hVal
			}
		}
		index++
	}

	log.Printf("Composed request obj: %s", req)
	return req
}
