package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Server listening on port 4221")

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Finished reading !!")
				break
			}
			fmt.Println("Error reading: ", err.Error())
			os.Exit(1)
		}

		url := strings.Split(string(buf[:n]), " ")[1]

		var res string
		if url == "/" {
			res = "HTTP/1.1 200 OK\r\n\r\n"
		} else if strings.Contains(url, "/echo/") {
			content := strings.Split(url, "/")[2]
			res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(content), content)
		} else {
			res = "HTTP/1.1 404 Not Found\r\n\r\n"
		}
		conn.Write([]byte(res))
	}

	fmt.Println("sending msg to client..")
}
