package shell

import (
	"context"
	"errors"
	"io"
	"time"

	httpapi "github.com/ipfs/go-ipfs-http-client"
)

var ErrNotSupported = errors.New("operation not supported")

// SetTimeout sets timeout for future requests. This method is not thread safe
func (s *Shell) SetTimeout(d time.Duration) error {
	if s.client == nil {
		return ErrNotSupported
	}

	s.client.Timeout = d

	api, err := httpapi.NewURLApiWithClient(s.url, s.client)
	if err != nil {
		return err
	}

	s.api = api
	return nil
}

func (s *Shell) http() (*httpapi.HttpApi, error) {
	api, ok := s.api.(*httpapi.HttpApi)
	if !ok {
		return nil, ErrNotSupported
	}
	return api, nil
}

func (s *Shell) Request(command string, args ...string) httpapi.RequestBuilder {
	http, err := s.http()
	if err != nil {
		return &errorRequestBuilder{err}
	}

	return http.Request(command, args...)
}

type errorRequestBuilder struct {
	error
}

func (b *errorRequestBuilder) Arguments(args ...string) httpapi.RequestBuilder {
	return b
}

func (b *errorRequestBuilder) BodyString(body string) httpapi.RequestBuilder {
	return b
}

func (b *errorRequestBuilder) BodyBytes(body []byte) httpapi.RequestBuilder {
	return b
}

func (b *errorRequestBuilder) Body(body io.Reader) httpapi.RequestBuilder {
	return b
}

func (b *errorRequestBuilder) FileBody(body io.Reader) httpapi.RequestBuilder {
	return b
}

func (b *errorRequestBuilder) Option(key string, value interface{}) httpapi.RequestBuilder {
	return b
}

func (b *errorRequestBuilder) Header(name, value string) httpapi.RequestBuilder {
	return b
}

func (b *errorRequestBuilder) Send(ctx context.Context) (*httpapi.Response, error) {
	return nil, b.error
}

func (b *errorRequestBuilder) Exec(ctx context.Context, res interface{}) error {
	return b.error
}

var _ httpapi.RequestBuilder = new(errorRequestBuilder)
