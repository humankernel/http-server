package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
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

	// debug persisten conn: sudo tcpdump -i any -n host localhost
	for {
		// read timeout to detech idle connections
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		// Read RequestLine (Method, Path, HTTP-version)
		requestLine, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return // client closed connection
			}
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

		// Read Headers
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

			headerParts := strings.SplitN(line, ":", 2)
			if len(headerParts) == 2 {
				headers[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
			}
		}

		// check for `Connection: Close` header
		connectionClose := false
		if connHeader, ok := headers["Connection"]; ok && strings.ToLower(connHeader) == "close" {
			connectionClose = true
		}

		// Read Body (for POST/UPDATE/DELETE)
		var reqBody string
		if method != "GET" && method != "HEAD" {
			contentLength := 0
			if cl, ok := headers["Content-Length"]; ok {
				contentLength, _ = strconv.Atoi(cl)
			}
			if contentLength > 0 {
				body := make([]byte, contentLength)
				if _, err := io.ReadFull(reader, body); err != nil {
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
			// save encodings into a map
			encodingsList := strings.Split(headers["Accept-Encoding"], ", ")
			encodings := make(map[string]bool)
			for _, v := range encodingsList {
				encodings[v] = true
			}
			if encodings["gzip"] {
				var buf bytes.Buffer
				gz := gzip.NewWriter(&buf)
				_, _ = gz.Write([]byte(body))
				gz.Close()
				gzipBody := buf.String()
				res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Encoding: gzip\r\nContent-Length: %d\r\n\r\n%v", len(gzipBody), gzipBody)
			} else {
				res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
			}

		case strings.HasPrefix(path, "/user-agent"):
			body := headers["User-Agent"]
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

		if connectionClose {
			return
		}
	}
}
