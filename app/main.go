package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	_"log"
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

	for {

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

		// parse headers once
		headerLines := strings.Split(req, "\r\n")[1:]
		headers := make(map[string]string)
		for _, line := range headerLines {
			if line == "" {
				break
			}
			kv := strings.SplitN(line, ":", 2)
			if len(kv) == 2 {
				headers[strings.ToLower(strings.TrimSpace(kv[0]))] = strings.TrimSpace(kv[1])
			}
		}

		if parts[0] == "GET" {
			response := "HTTP/1.1 404 Not Found\r\n\r\n"

			if len(parts) >= 2 {
				path := parts[1]

				if path == "/" {
					response = "HTTP/1.1 200 OK\r\n\r\n"

				} else if strings.HasPrefix(path, "/echo/") {
					echoRaw := path[len("/echo/"):]
					echoStr, _ := url.PathUnescape(echoRaw)

					accept := strings.ToLower(headers["accept-encoding"])

					if strings.Contains(accept, "gzip") {
						var buf bytes.Buffer
						gz := gzip.NewWriter(&buf)
						gz.Write([]byte(echoStr))
						gz.Close()

						compressed := buf.Bytes()

						response := fmt.Sprintf(
							"HTTP/1.1 200 OK\r\n"+
								"Content-Encoding: gzip\r\n"+
								"Content-Type: text/plain\r\n"+
								"Content-Length: %d\r\n\r\n",
							len(compressed),
						)

						conn.Write([]byte(response))
						conn.Write(compressed)

						if strings.ToLower(headers["connection"]) == "close" {
							return
						}
						continue
					}

					length := len(echoStr)
					response := fmt.Sprintf(
						"HTTP/1.1 200 OK\r\n"+
							"Content-Type: text/plain\r\n"+
							"Content-Length: %d\r\n\r\n%s",
						length, echoStr,
					)

					conn.Write([]byte(response))

					if strings.ToLower(headers["connection"]) == "close" {
						return
					}
					continue
				}

				// /user-agent
				if path == "/user-agent" || strings.HasPrefix(path, "/user-agent/") {
					echoRaw := headers["user-agent"]
					echoStr, _ := url.PathUnescape(echoRaw)
					length := len([]byte(echoStr))
					response = fmt.Sprintf(
						"HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
						length, echoStr,
					)
				}

				// /files
				if strings.HasPrefix(path, "/files/") {
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
								length, string(fileBytes),
							)
						}
					}
				}
			}

			_, werr := conn.Write([]byte(response))
			if werr != nil {
				fmt.Println("Error writing response:", werr)
			}

			if strings.ToLower(headers["connection"]) == "close" {
				return
			}
		}

		if parts[0] == "POST" {
			response := "HTTP/1.1 201 Created\r\n\r\n"

			if len(parts) >= 2 {
				lines := strings.Split(req, "\r\n")
				path := strings.Split(lines[0], " ")

				foldPath := strings.Split(path[1], "/")
				filename := foldPath[2]

				var contentLength int
				for _, line := range lines {
					if strings.HasPrefix(line, "Content-Length:") {
						fmt.Sscanf(line, "Content-Length: %d", &contentLength)
					}
				}

				partsReq := strings.SplitN(req, "\r\n\r\n", 2)
				body := ""
				if len(partsReq) == 2 {
					body = partsReq[1]
				}

				if len(body) > contentLength {
					body = body[:contentLength]
				}

				baseDir := os.Args[2]
				filePath := filepath.Join(baseDir, filename)

				os.WriteFile(filePath, []byte(body), 0644)
			}

			conn.Write([]byte(response))

			if strings.ToLower(headers["connection"]) == "close" {
				return
			}
		}
	}
}
