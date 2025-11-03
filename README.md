# ⚙️ Simple HTTP Server from Scratch (Go)

A minimal yet fully functional **HTTP/1.1 server built from scratch in Go**, without using `net/http` for handling requests or responses.
This project was inspired by the [Codecrafters HTTP Server Challenge](https://app.codecrafters.io/courses/http-server) and rebuilt from the ground up to deepen understanding of low-level networking and the HTTP protocol.

## 🚀 Features

✅ **Manual TCP Handling** — Uses Go’s `net` package directly (`net.Listen`, `net.Conn`) to accept and manage client connections.

✅ **HTTP/1.1 Parsing** — Parses request lines, headers, and bodies manually via buffered I/O.

✅ **Persistent Connections** — Supports `Connection: keep-alive` and `Connection: close`.

✅ **Custom Routing** — Handles multiple routes without using frameworks:

* `/` — Simple health check.
* `/echo/<text>` — Responds with the given text.
* `/user-agent` — Returns the client’s `User-Agent` header.
* `/files/<filename>` —
  • `GET`: Serves static files from a provided directory.
  • `POST`: Writes file contents to disk.

✅ **Gzip Compression** — Compresses responses when the client includes `Accept-Encoding: gzip`.

✅ **File I/O Handling** — Secure path resolution and support for reading/writing binary files.

✅ **Proper Status Codes** — Returns `200 OK`, `201 Created`, `404 Not Found`, `500 Internal Server Error`, etc.

✅ **Concurrent Connections** — Handles multiple clients via goroutines.

✅ **Verbose Logging** — Logs request lifecycle and connection details.

## 🧩 Getting Started

### 1. Clone the repo

```bash
git clone https://github.com/humankernel/http-server.git
cd http-server
```

### 2. Run the server

You must provide a file directory path for `/files` route:

```bash
go run app/main.go ./tmp
```

### 3. Try some requests

```bash
# Echo
curl -v http://localhost:4221/echo/hello

# User-Agent
curl -v http://localhost:4221/user-agent

# Upload a file
curl -v -X POST --data-binary @example.txt http://localhost:4221/files/example.txt

# Download a file
curl -v http://localhost:4221/files/example.txt
```

### 4. Observe raw traffic (optional)

```bash
sudo tcpdump -i any -n host localhost
```

---

### 📂 Example Response

```
> GET /echo/hi HTTP/1.1
> Host: localhost:4221

< HTTP/1.1 200 OK
< Content-Type: text/plain
< Content-Length: 2

hi
```

## 🧭 Project Structure

```
http-server/
├── app/
│   └── main.go        # Core server implementation
├── tmp/               # Example directory for /files route
└── README.md
```

## 💡 Future Improvements

* Add proper routing abstraction (similar to `http.ServeMux`)
* Implement chunked transfer encoding and streaming responses
* Improve security and header handling