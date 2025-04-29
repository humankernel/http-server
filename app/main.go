package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		log.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			// Continue accepting connections instead of exiting
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// Read Request
		method, path, reqHeaders, reqBody, err := readRequest(reader)
		if err != nil {
			if err == io.EOF {
				log.Printf("Client closed connection")
			} else {
				log.Printf("Error reading request: %v", err)
			}
			return
		}

		closeConnection := false
		if connHeader, ok := reqHeaders["Connection"]; ok && strings.ToLower(connHeader) == "close" {
			closeConnection = true
		}

		// Handle Request
		response := handleRequest(method, path, reqHeaders, reqBody)

		// Send response
		if _, err = conn.Write([]byte(response)); err != nil {
			log.Println("Error writing response:", err)
			return
		}
		log.Println("Response successfully sent to client.")

		if closeConnection {
			return
		}
	}
}

func readRequest(reader *bufio.Reader) (method, path string, headers map[string]string, body []byte, err error) {
	// Read request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return "", "", nil, nil, err
	}

	parts := strings.Split(strings.TrimSpace(requestLine), " ")
	method, path = parts[0], parts[1]

	// Read Headers
	headers = make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil || strings.TrimSpace(line) == "" {
			break // End of headers
		}

		headerParts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(headerParts) == 2 {
			key := strings.TrimSpace(headerParts[0])
			value := strings.TrimSpace(headerParts[1])
			headers[key] = value
		}
	}

	// Read Body
	if contentLengthStr, ok := headers["Content-Length"]; ok {
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return "", "", nil, nil, fmt.Errorf("invalid Content-Length")
		}

		if contentLength > 0 {
			body = make([]byte, contentLength)
			if _, err = io.ReadFull(reader, body); err != nil {
				return "", "", nil, nil, fmt.Errorf("error reading body: %s", err)
			}
		}
	}

	return method, path, headers, body, nil
}

func handleRequest(method, path string, headers map[string]string, body []byte) (res string) {
	responseHeaders := make(map[string]string)
	responseHeaders["Connection"] = headers["Connection"]

	status := http.StatusNotFound
	var responseBody []byte

	switch {
	case path == "/":
		status = http.StatusOK
		responseHeaders["Content-Length"] = "0"

	case strings.HasPrefix(path, "/user-agent"):
		status = http.StatusOK
		responseBody = []byte(headers["User-Agent"])
		responseHeaders["Content-Type"] = "text/plain"
		responseHeaders["Content-Length"] = fmt.Sprintf("%v", len(responseBody))

	case strings.HasPrefix(path, "/echo/"):
		status = http.StatusOK
		responseBody = []byte(strings.TrimPrefix(path, "/echo/"))
		responseHeaders["Content-Type"] = "text/plain"
		responseHeaders["Content-Length"] = fmt.Sprintf("%v", len(responseBody))

	case strings.HasPrefix(path, "/files"):
		fileDir := os.Args[2]
		fileName := filepath.Clean(strings.TrimPrefix(path, "/files/"))
		filePath, _ := filepath.Abs(filepath.Join(fileDir, fileName))

		switch method {
		case "GET":
			content, err := os.ReadFile(filePath)
			if err != nil {
				status = http.StatusNotFound
			} else {
				status = http.StatusOK
				responseBody = content
				responseHeaders["Content-Length"] = fmt.Sprintf("%v", len(responseBody))
				responseHeaders["Content-Type"] = "application/octet-stream"
			}
		case "POST":
			err := os.WriteFile(filePath, body, 0644)
			if err != nil {
				status = http.StatusInternalServerError
			} else {
				status = http.StatusCreated
				responseHeaders["Content-Length"] = "0"
			}

		default:
			status = http.StatusNotImplemented
		}
	}

	if strings.Contains(headers["Accept-Encoding"], "gzip") {
		responseHeaders["Content-Encoding"] = "gzip"
		responseBody = compressGzip(responseBody)
		responseHeaders["Content-Length"] = fmt.Sprintf("%v", len(responseBody))
	}

	return generateResponse(responseHeaders, status, responseBody)
}

func generateResponse(headers map[string]string, status int, body []byte) string {
	var res strings.Builder

	// status Line
	res.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", status, http.StatusText(status)))

	// headers
	if len(body) > 0 {
		headers["Content-Length"] = strconv.Itoa(len(body))
	}
	for key, value := range headers {
		if value != "" {
			res.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		}
	}
	res.WriteString("\r\n")

	// body
	if len(body) > 0 {
		res.Write(body)
	}

	return res.String()
}

func compressGzip(content []byte) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write(content)
	gz.Close()
	return buf.Bytes()
}
