package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Server listening on port 4221")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue // continue accepting conections
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Read RequestLine (Method, Path, HTTP-version)
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request data: ", err.Error())
		return
	}
	requestLine = strings.TrimSpace(requestLine)
	reqItems := strings.Split(requestLine, " ")
	if len(reqItems) < 2 {
		fmt.Println("Malformed request line: ", requestLine)
		return
	}
	method, path := reqItems[0], reqItems[1]

	// Read Headers (one by one)
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading headers: ", err)
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break // End of headers
		}

		// headers are typically in key:value format
		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			headers[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
		}
	}
	userAgent := headers["User-Agent"]
	acceptEncoding := headers["Accept-Encoding"]

	// Read Body (for POST/UPDATE/DELETE)
	var reqBody string
	if method != "GET" {
		var contentLength int
		if v, ok := headers["Content-Length"]; ok {
			contentLength, _ = strconv.Atoi(v)
		}
		if contentLength > 0 {
			body := make([]byte, contentLength)
			if _, err := reader.Read(body); err != nil {
				fmt.Println("Error reading body", err)
				return
			}
			reqBody = string(body)
		}
	}

	// Create Response
	var res string
	switch {
	case path == "/":
		res = "HTTP/1.1 200 OK\r\n\r\n"
	case strings.HasPrefix(path, "/echo/"):
		body := strings.TrimPrefix(path, "/echo/")
		if acceptEncoding == "gzip" {
			res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Encoding: gzip\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
		} else {
			res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
		}
	case strings.HasPrefix(path, "/user-agent"):
		body := userAgent
		res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
	case strings.HasPrefix(path, "/files"):
		dir := os.Args[2]
		filename := strings.TrimPrefix(path, "/files/")
		switch method {
		case "GET":
			dat, err := os.ReadFile(dir + filename)
			if err != nil {
				res = "HTTP/1.1 404 Not Found\r\n\r\n"
			} else {
				res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(dat), dat)
			}
		case "POST":
			err := os.WriteFile(dir+filename, []byte(reqBody), 0644)
			if err != nil {
				res = "HTTP/1.1 404 Not Found\r\n\r\n"
			} else {
				res = "HTTP/1.1 201 Created\r\n\r\n"
			}
		}
	default:
		res = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	// Write the res to the client
	if _, err = conn.Write([]byte(res)); err != nil {
		fmt.Println("Error writing response: ", err)
		return
	}

	fmt.Println("Response succesfully sended to client.")
}
