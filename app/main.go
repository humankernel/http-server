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
		msg := buf[:n]
		elements := strings.Split(string(msg), " ")
		if len(elements) >= 1 {
			url := elements[1]
			fmt.Println(url)

			switch url {
			case "/":
				conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
			default:
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			}
		}
	}

	fmt.Println("sending msg to client..")
}
