# Simple HTTP-Server from Scratch

Change from [codecrafters/http-server](https://app.codecrafters.io/courses/http-server)

See also the other branches :)

## Todo 

- Improve security (dont use req headers as res headers)
- Use a better way to create handlers for the routes
- Use streaming for files

## Getting Started

1. Clone: `git clone https://github.com/humankernel/http-server/tree/main
2. Run server: `go run app/main.go`
3. Use curl: `curl -v http://localhost:4221/echo/abc`
4. Debug: `sudo tcpdump -i any -n host localhost`