package myhttp

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"net"
	"slices"
	"strings"
)

type Server struct {
	Name         string
	Network      string
	Address      string
	handlers     map[string]func(Request, *Response)
	wildPatterns []string
	listener     net.Listener
}

func NewTCPServer(address string) *Server {
	const defaultServerName string = "myhttp/0.0.1"
	return &Server{
		Name:         defaultServerName,
		Network:      "tcp",
		Address:      address,
		handlers:     map[string]func(Request, *Response){},
		wildPatterns: make([]string, 0),
	}
}

func (s *Server) SetHandler(pattern string, handler func(Request, *Response)) error {
	if _, ok := s.handlers[pattern]; ok {
		return errors.New("Pattern already set")
	}
	s.handlers[pattern] = handler

	// Register wildcard pattern
	if len(pattern) > 0 && pattern[len(pattern)-1:] == "/" {
		s.wildPatterns = append(s.wildPatterns, pattern)
		slices.SortFunc(s.wildPatterns, func(a, b string) int {
			return cmp.Compare(len(b), len(a)) // From longest to shortest string
		})
	}
	return nil
}

// You shouldn't need this
func (s *Server) UnsetHandler(pattern string) {
	delete(s.handlers, pattern)
	for i, p := range s.wildPatterns {
		if p == pattern {
			s.wildPatterns = append(s.wildPatterns[:i], s.wildPatterns[i+1:]...)
		}
	}
}

func (s *Server) ListenAndServe() error {
	listener, err := net.Listen(s.Network, s.Address)
	if err != nil {
		return err
	}
	s.listener = listener

	s.serve() // TODO make goroutine and wait for os signal after for graceful shutdown
	return nil
}

func (s *Server) serve() {
	defer s.listener.Close()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// TODO: use decent logging here?
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	request, err := ReadRequest(conn)
	if err != nil {
		if err != io.EOF {
			// TODO: use decent logging here?
			fmt.Println("Error reading request: ", err.Error())
		}
		conn.Close()
		return
	}

	handler := s.findHandler(request)
	response := s.InitResponse()
	if val, ok := request.Headers["Connection"]; ok && val == "close" {
		response.Headers["Connection"] = "close"
	}
	handler(*request, response)

	// Prep, check, and send
	response.Prep()
	err = SendResponse(conn, response)
	if err != nil {
		fmt.Println("Failed to send response on connection", err.Error())
		return
	}

	// Recurse to handle keep-alive
	connectionHeader, ok := response.Headers["Connection"]
	if !ok || connectionHeader == "keep-alive" {
		s.handleConn(conn)
	}
}

// A default response
func (s *Server) InitResponse() *Response {
	response := &Response{
		HTTPVersion:  "HTTP/1.1",
		StatusCode:   200,
		ReasonPhrase: "OK",
	}
	response.Headers = map[string]string{
		"Content-Type":   "text/html; charset=utf-8",
		"Connection":     "keep-alive",
		"Server":         s.Name,
		"Content-Length": "0",
	}
	return response
}

func (s *Server) findHandler(request *Request) func(Request, *Response) {
	// Exact matches always first
	handler, ok := s.handlers[request.Path]
	if ok {
		return handler
	}

	// Then wildcard matches
	for _, pattern := range s.wildPatterns {
		if strings.HasPrefix(request.Path, pattern) {
			return s.handlers[pattern] // BOLDY ASSUME ITS IN THE MAP
		}
	}

	return Default404Handler
}

func Default404Handler(request Request, response *Response) {
	response.HTTPVersion = "HTTP/1.1"
	response.StatusCode = 404
	response.ReasonPhrase = "Not Found"
	response.Headers = map[string]string{
		"Content-Length": "0",
	}
}

func SendResponse(conn net.Conn, response *Response) error {
	_, err := conn.Write([]byte(response.String()))
	if err != nil {
		return err
	}
	return nil
}

// func hello(w http.ResponseWriter, req *http.Request) {
// 	// fmt.Fprintf(w, "hello")
// 	return
// }

// func main() {
// 	http.HandleFunc("/hello", hello)
// 	http.ListenAndServe(":4221", nil)
// }

// func main() {
// 	l, err := net.Listen("tcp", "0.0.0.0:4221")
// 	// http.ListenAndServe()
// 	if err != nil {
// 		fmt.Println("Failed to bind to port 4221")
// 		os.Exit(1)
// 	}
// 	for {
// 		conn, err := l.Accept()
// 		if err != nil {
// 			fmt.Println("Error accepting connection: ", err.Error())
// 			continue
// 		}
// 		go HandleConn(conn)
// 		// conn.Close()
// 		// _, err = conn.Write([]byte("d"))
// 		// conn.Close()
// 	}
// }

// func HandleConn(conn net.Conn) {
// 	defer conn.Close()
// 	request, err := ReadRequest(conn)
// 	if err != nil {
// 		if err != io.EOF {
// 			fmt.Println("Error reading request: ", err.Error())
// 		}
// 		conn.Close()
// 		return
// 	}

// 	response := new(Response)
// 	response.HTTPVersion = "HTTP/1.1"
// 	// if connectionHeader, ok := request.Headers["Connection"]; ok {
// 	// 	if connectionHeader != "keep-alive" {
// 	// 		defer conn.Close()
// 	// 	}
// 	// }
// 	connectionHeader, ok := request.Headers["Connection"]
// 	// if ok && connectionHeader == "close" {
// 	// 	defer conn.Close()
// 	// } else {
// 	// 	defer HandleConn(conn)
// 	// }
// 	if !ok || connectionHeader == "keep-alive" {
// 		defer HandleConn(conn)
// 	}

// 	if request.Path == "/" {
// 		response.StatusCode = 200
// 		response.ReasonPhrase = "OK"
// 		response.Headers = map[string]string{
// 			"Content-Type":                     "text/html; charset=utf-8",
// 			"Access-Control-Allow-Origin":      "*",
// 			"Access-Control-Allow-Credentials": "true",
// 			"Server":                           "flocus/0.0.1",
// 		}
// 	} else if strings.HasPrefix(request.Path, "/echo") {
// 		response.StatusCode = 200
// 		response.ReasonPhrase = "OK"
// 		response.Headers = map[string]string{
// 			"Content-Type": "text/html; charset=utf-8",
// 		}
// 		response.Body = request.Path[6:]
// 		// response.Body = "wtf"
// 		// response.Body = request.Headers["User-Agent"]
// 		// fmt.Println(request.Headers["User-Agent"])
// 		// fmt.Println(request.Path[6:])
// 	} else if request.Path == "/user-agent" {
// 		response.StatusCode = 200
// 		response.ReasonPhrase = "OK"
// 		response.Headers = map[string]string{
// 			"Content-Type": "text/html; charset=utf-8",
// 		}
// 		response.Body = request.Headers["User-Agent"]
// 	} else if request.Path == "/wtf" {
// 		response.StatusCode = 200
// 		response.ReasonPhrase = "OK"
// 		response.Headers = map[string]string{
// 			"Content-Type": "text/html; charset=utf-8",
// 		}
// 		response.Body = request.Headers["User-Agent"]
// 	} else {
// 		response.StatusCode = 404
// 		response.ReasonPhrase = "Not Found"
// 	}

// 	if !response.Ready() {
// 		fmt.Println("Response was malformed right before sending...")
// 		return
// 	}
// 	_, err = conn.Write([]byte(response.String()))
// 	// _, err = conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"))
// 	if err != nil {
// 		fmt.Println("Error writing response: ", err.Error())
// 		return
// 	}
// }
