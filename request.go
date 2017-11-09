package kaleidoscope

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Request struct {
	ApiBase string
	Command string
	Args    []string
	Opts    map[string]string
	Body    io.Reader
	Headers map[string]string
}

type RequestOptions map[string]string

func NewRequest(url, command string, opts RequestOptions, args ...string) *Request {
	req := newRequest(url, command, args...)
	for k, v := range opts {
		req.Opts[k] = v
	}
	return req
}

func newRequest(url, command string, args ...string) *Request {
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	opts := map[string]string{
		"encoding":        "json",
		"stream-channels": "true",
	}
	return &Request{
		ApiBase: url + "/api/v0",
		Command: command,
		Args:    args,
		Opts:    opts,
		Headers: make(map[string]string),
	}
}

func (r Request) Send(c http.Client) (*Response, error) {
	url := r.getURL()

	req, err := http.NewRequest("POST", url, r.Body)
	if err != nil {
		return nil, err
	}

	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	contentType := resp.Header.Get("Content-Type")
	parts := strings.Split(contentType, ";")
	contentType = parts[0]

	nresp := new(Response)

	nresp.Output = resp.Body
	if resp.StatusCode >= http.StatusBadRequest {
		e := &Error{
			Command: r.Command,
		}
		switch {
		case resp.StatusCode == http.StatusNotFound:
			e.Message = "command not found"
		case contentType == "text/plain":
			if out, err := ioutil.ReadAll(resp.Body); err != nil {
				e.Message = err.Error()
			} else {
				e.Message = string(out)
			}
		case contentType == "application/json":
			if err = json.NewDecoder(resp.Body).Decode(e); err != nil {
				e.Message = err.Error()
			}
		default:
			if out, err := ioutil.ReadAll(resp.Body); err != nil {
				e.Message = err.Error()
			} else {
				e.Message = fmt.Sprintf("unknown error: %q - %q", contentType, out)
			}
		}
		nresp.Error = e
		nresp.Output = nil

		ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}

	return nresp, nil
}

func (r *Request) getURL() string {

	values := make(url.Values)
	for _, arg := range r.Args {
		values.Add("arg", arg)
	}
	for k, v := range r.Opts {
		values.Add(k, v)
	}

	return fmt.Sprintf("%s/%s?%s", r.ApiBase, r.Command, values.Encode())
}

type Response struct {
	Output io.ReadCloser
	Error  *Error
}

func (r *Response) Close() error {
	if r.Output != nil {
		// always drain output (response body)
		ioutil.ReadAll(r.Output)
		return r.Output.Close()
	}
	return nil
}

type Error struct {
	Command string
	Message string
	Code    int
}

func (e *Error) Error() string {
	var out string
	if e.Command != "" {
		out = e.Command + ": "
	}
	if e.Code != 0 {
		out = fmt.Sprintf("%s%d: ", out, e.Code)
	}
	return out + e.Message
}
