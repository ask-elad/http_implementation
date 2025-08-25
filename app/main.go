package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var (
	_ = net.Listen
	_ = os.Exit
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	// ensure connection is closed
	defer conn.Close()

	// read request into buffer
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	// use only the bytes actually read
	req := string(buffer[:n])

	// get the first line (request line) ending at CRLF
	idx := strings.Index(req, "\r\n")
	var firstLine string
	if idx == -1 {
		firstLine = req
	} else {
		firstLine = req[:idx]
	}

	// split the request line into fields: METHOD PATH VERSION
	parts := strings.Fields(firstLine)

	// debug prints (optional)
	fmt.Println(n)
	fmt.Println(parts)

	// default to 404 if malformed
	response := "HTTP/1.1 404 Not Found\r\n\r\n"
	if len(parts) >= 2 {
		path := parts[1]

		// handle root path first (stage IA4)
		if path == "/" {
			response = "HTTP/1.1 200 OK\r\n\r\n"
		} else if strings.HasPrefix(path, "/echo/") {
			echoRaw := path[len("/echo/"):]
			echoStr, _ := url.PathUnescape(echoRaw)
			length := len([]byte(echoStr))

			response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", length, echoStr)
		} else if strings.HasPrefix(path, "/user-agent/") {

			headerLines := strings.Split(req, "\r\n")[1:] // skip first line
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
		}
	}

	// write the chosen response
	_, werr := conn.Write([]byte(response))
	if werr != nil {
		fmt.Println("Error writing response:", werr)
	}
}

