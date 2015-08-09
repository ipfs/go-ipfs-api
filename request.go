package shell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	files "github.com/whyrusleeping/go-multipart-files"
)

type Request struct {
	ApiBase string
	Command string
	Args    []string
	Opts    map[string]string
	Body    io.Reader
	Headers map[string]string
}

func NewRequest(url, command string, args ...string) *Request {
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
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func (r *Request) Send(c *http.Client) (*Response, error) {
	url := r.getURL()

	req, err := http.NewRequest("POST", url, r.Body)
	if err != nil {
		return nil, err
	}

	if fr, ok := r.Body.(*files.MultiFileReader); ok {
		req.Header.Set("Content-Type", "multipart/form-data; boundary="+fr.Boundary())
		req.Header.Set("Content-Disposition", "form-data: name=\"files\"")
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
		switch {
		case resp.StatusCode == http.StatusNotFound:
			nresp.Error = &Error{"command not found"}
		case contentType == "text/plain":
			out, err := ioutil.ReadAll(resp.Body)
			_ = err // already in an error case...
			e := &Error{string(out)}
			nresp.Error = e
		default:
			buf := new(bytes.Buffer)
			io.Copy(buf, resp.Body)
			fmt.Println(contentType)
			fmt.Println(buf.String())
			e := new(Error)
			err := json.NewDecoder(buf).Decode(e)
			_ = err
			nresp.Error = e
		}
		nresp.Output = nil

		// drain body and close
		ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}

	return nresp, nil
}

func (r *Request) getURL() string {
	argstring := ""
	for _, arg := range r.Args {
		argstring += fmt.Sprintf("arg=%s&", arg)
	}
	for k, v := range r.Opts {
		argstring += fmt.Sprintf("%s=%s&", k, v)
	}

	return fmt.Sprintf("%s/%s?%s", r.ApiBase, r.Command, argstring)
}
