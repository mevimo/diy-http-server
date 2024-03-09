package myhttp

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	HTTPVersion  string
	StatusCode   uint16
	ReasonPhrase string
	Headers      map[string]string
	Body         string
}

// Returns a valid and whole HTTP response.
func (r *Response) String() string {
	const CRLF string = "\r\n"
	var sb strings.Builder

	sb.WriteString(r.HTTPVersion)
	sb.WriteString(" ")
	sb.WriteString(strconv.FormatUint(uint64(r.StatusCode), 10))
	sb.WriteString(" ")
	sb.WriteString(r.ReasonPhrase)
	sb.WriteString(CRLF)

	for k, v := range r.Headers {
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(v)
		sb.WriteString(CRLF)
	}
	sb.WriteString(CRLF)
	sb.WriteString(r.Body)
	return sb.String()
}

func (r *Response) Prep() {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	if _, ok := r.Headers["Date"]; !ok {
		r.Headers["Date"] = time.Now().Format(time.RFC1123)
	}
	r.Headers["Content-Length"] = strconv.Itoa(len(r.Body))
}

// Checks if all required fields are set to send out the response: `HTTPVersion`,
// `StatusCode`, and `ReasonPhrase`.
func (r *Response) EnsureReady() error {
	if len(r.HTTPVersion) == 0 {
		return errors.New("Response missing HTTP version.")
	}
	if r.StatusCode == 0 {
		return errors.New("Response missing HTTP status code.")
	}
	if len(r.ReasonPhrase) == 0 {
		return errors.New("Response missing HTTP reasonphrase.")
	}
	return nil
}
