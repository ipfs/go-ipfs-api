package shell

import "net/http"

type transport struct {
	token  string
	httptr http.RoundTripper
}

func newAuthenticatedTransport(tr http.RoundTripper, token string) *transport {
	return &transport{
		token:  token,
		httptr: tr,
	}
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.httptr.RoundTrip(req)
}
