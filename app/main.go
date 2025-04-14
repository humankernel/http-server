package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
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

	fmt.Println("Successfully accepted incoming conection.")

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Read RequestLine (method, path, http-version)
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request data: ", err.Error())
		return
	}
	req := strings.TrimSpace(requestLine)
	reqItems := strings.Split(req, " ")
	if len(reqItems) < 2 {
		fmt.Println("Malformed request line: ", requestLine)
		return
	}
	method, path := reqItems[0], reqItems[1]

	fmt.Println(method, path)

	// Read headers one by one
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

	res := ""

	switch {
	case path == "/":
		res = "HTTP/1.1 200 OK\r\n\r\n"
	case strings.HasPrefix(path, "/echo/"):
		body := path[len("/echo/"):]
		res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
	case strings.HasPrefix(path, "/user-agent"):
		body := userAgent
		res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
	default:
		res = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	// Write the res to the client
	_, err = conn.Write([]byte(res))
	if err != nil {
		fmt.Println("Error writing response: ", err)
		return
	}

	fmt.Println("Response sent to client.")
}
