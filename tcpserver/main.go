package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Start TCP Server")
	receiveTCPConnection(listener)
}

func receiveTCPConnection(listener *net.TCPListener) {
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Fatal(err)
		}

		go httpHandler(conn)
	}
}
func httpHandler(conn *net.TCPConn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	var path string
	header := make(map[string]string)
	isFirst := true

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}

		if isFirst {
			isFirst = false
			headerLine := strings.Fields(line)
			header["Method"] = headerLine[0]
			header["Path"] = headerLine[1]
			continue
		}

		// Header Fields
		headerFields := strings.SplitN(line, ": ", 2)
		header[headerFields[0]] = headerFields[1]
	}
	path = header["Path"]

	// non-EOF error がある場合
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	// リクエストボディ
	s := header["Content-Length"]
	if s == "" {
		s = "0"
	}
	contentLength, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	buf := make([]byte, contentLength)
	if _, err := io.ReadFull(conn, buf); err != nil {
		panic(err)
	}

	if path == "/style.css" {
		time.Sleep(500 * time.Millisecond)

		body := `h1 { color:blue; font-weight:bold; }`

		conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
		conn.Write([]byte(fmt.Sprintf("Content-Length: %d\r\n", len(body))))
		conn.Write([]byte("Content-Type: text/css; charset=utf-8\r\n"))
		conn.Write([]byte("\r\n"))
		conn.Write([]byte(body))
		return
	}

	fmt.Println("BODY length:", len(buf))

	conn.Write([]byte("HTTP/1.1 103 Early Hints\r\n"))
	conn.Write([]byte("Link: /styles.css; rel=preload; as=style\r\n"))
	conn.Write([]byte("\r\n"))
	conn.Write([]byte("\r\n"))


	time.Sleep(1000 * time.Millisecond)

	body := `<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
    <link rel="stylesheet" type="text/css" href="/style.css">
    <title>Resource Hints test</title>
  </head>
  <body>
    <h1>Resource Hints test</h1>
    <p>css preload</p>
  </body>
</html>
`


	conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
	conn.Write([]byte(fmt.Sprintf("Content-Length: %d\r\n", len(body))))
	conn.Write([]byte("Content-Type: text/html; charset=utf-8\r\n"))
	conn.Write([]byte("\r\n"))
	conn.Write([]byte(body))

	/*

	--enable-features=EarlyHintsPreloadForNavigation

	performance.getEntriesByName('http://localhost:8080/style.css')[0].initiatorType
	'link'
	 */
}
