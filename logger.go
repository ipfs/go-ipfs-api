package shell

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Logger is used to handle incoming logs from the ipfs node
type Logger struct {
	r io.ReadCloser
}

// Next is used to retrieve the next event from the logging system
func (l Logger) Next() (map[string]interface{}, error) {
	var out map[string]interface{}
	if err := json.NewDecoder(l.r).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// Close is used to close our reader
func (l Logger) Close() error {
	return l.r.Close()
}

// GetLogs is used to retrieve a parsable logger object
func (s *Shell) GetLogs() (Logger, error) {
	logURL := fmt.Sprintf("http://%s/api/v0/log/tail", s.url)
	req, err := http.NewRequest("GET", logURL, nil)
	if err != nil {
		return Logger{}, err
	}
	hc := &http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return Logger{}, err
	}
	return Logger{resp.Body}, nil
}
