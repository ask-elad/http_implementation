package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var (
	_ = net.Listen
	_ = os.Exit
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	// Listen on port 4221 (required by CodeCrafters)
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	req := string(buffer[:n])

	// get the first line (request line)
	idx := strings.Index(req, "\r\n")
	firstLine := req
	if idx != -1 {
		firstLine = req[:idx]
	}

	parts := strings.Fields(firstLine)

	// debug prints (optional)
	fmt.Println(n)
	fmt.Println(parts)

	response := "HTTP/1.1 404 Not Found\r\n\r\n"

	if len(parts) >= 2 {
		path := parts[1]

		// root path
		if path == "/" {
			response = "HTTP/1.1 200 OK\r\n\r\n"
		} else if strings.HasPrefix(path, "/echo/") {
			echoRaw := path[len("/echo/"):]
			echoStr, _ := url.PathUnescape(echoRaw)
			length := len([]byte(echoStr))
			response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", length, echoStr)
		} else if path == "/user-agent" || strings.HasPrefix(path, "/user-agent/") {
			// parse headers
			headerLines := strings.Split(req, "\r\n")[1:] // skip request line
			headers := make(map[string]string)
			for _, line := range headerLines {
				if line == "" {
					break
				}
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					headers[strings.ToLower(strings.TrimSpace(parts[0]))] = strings.TrimSpace(parts[1])
				}
			}

			echoRaw := headers["user-agent"]
			echoStr, _ := url.PathUnescape(echoRaw)
			length := len([]byte(echoStr))
			response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", length, echoStr)
		} else if strings.HasPrefix(path, "/files/") {
			fileName := path[len("/files/"):]
			dirName := os.Args[2]
			pathName := filepath.Join(dirName, fileName)

			_, err := os.Stat(pathName)
			if err != nil {
				response = "HTTP/1.1 404 Not Found\r\n\r\n"
			} else {
				fileBytes, err := os.ReadFile(pathName)
				if err != nil {
					response = "HTTP/1.1 500 Internal Server Error\r\n\r\n"
				} else {
					length := len(fileBytes)
					response = fmt.Sprintf(
						"HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s",
						length,
						string(fileBytes),
					)
				}
			}
		}

	}

	_, werr := conn.Write([]byte(response))
	if werr != nil {
		fmt.Println("Error writing response:", werr)
	}
}
