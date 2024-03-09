package myhttp

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
)

type Request struct {
	Conn        net.Conn
	Method      string
	Path        string
	HTTPVersion string
	Body        string
	Headers     map[string]string
}

func ReadRequest(conn net.Conn) (req *Request, err error) {
	req = new(Request)
	req.Conn = conn
	reader := bufio.NewReader(conn)

	err = parseHTTPStartLine(reader, req)
	if err != nil {
		return nil, err
	}
	req.Headers, err = parseHTTPHeaders(reader)
	if err != nil {
		return nil, err
	}

	// Parse request body if Content-Length header is set
	if contentLength, ok := req.Headers["Content-Length"]; ok {
		contentLength, err := strconv.Atoi(contentLength)
		if err != nil {
			return nil, err
		}
		buf := make([]byte, contentLength)
		_, err = io.ReadFull(reader, buf)
		if err != nil {
			return nil, err
		}
		req.Body = string(buf)
	}

	return req, nil
}

func parseHTTPStartLine(reader *bufio.Reader, req *Request) error {
	startLine, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	startLineParts := strings.Split(strings.TrimSpace(startLine), " ")
	if len(startLineParts) != 3 {
		return errors.New("Failed to find 3 parts in HTTP start line.")
	}
	req.Method = startLineParts[0]
	req.Path = startLineParts[1]
	req.HTTPVersion = startLineParts[2]
	return nil
}

// Will start looking for HTTP headers from where the reader has advanced so far
func parseHTTPHeaders(reader *bufio.Reader) (headers map[string]string, err error) {
	headers = map[string]string{}
	for {
		line, err := reader.ReadString('\n')
		if len(line) == 0 && err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// Handle both \r\n
		if line[len(line)-2] == '\r' {
			line = line[:len(line)-2]
		} else {
			line = line[:len(line)-1]
		}

		// End of headers
		if line == "" {
			break
		}

		headerLine := strings.Split(line, ": ")
		headerName := headerLine[0]
		headerValue := headerLine[1]
		headers[headerName] = headerValue
	}
	return headers, nil
}
