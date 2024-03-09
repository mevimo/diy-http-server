package main

import "github.com/codecrafters-io/http-server-starter-go/app/myhttp"

func main() {
	server := myhttp.NewTCPServer("localhost:4221")
	server.SetHandler("/", RootHandler)
	server.SetHandler("/user-agent", UserAgentHandler)
	server.SetHandler("/echo/", EchoHandler)

	err := server.ListenAndServe()
	if err != nil {
		panic("Could not start server :(")
	}
}

func RootHandler(req myhttp.Request, res *myhttp.Response) {
	if req.Path != "/" {
		myhttp.Default404Handler(req, res)
	}
} // Just a 200 OK

func UserAgentHandler(req myhttp.Request, res *myhttp.Response) {
	res.Body = req.Headers["User-Agent"]
}

func EchoHandler(req myhttp.Request, res *myhttp.Response) {
	res.Body = req.Path[6:] // everything after the echo
}
