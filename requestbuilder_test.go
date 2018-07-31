package shell

import (
	"testing"
	"time"

	"github.com/cheekybits/is"
)

func TestRequestBuilder(t *testing.T) {
	is := is.New(t)

	now := time.Now()
	r := RequestBuilder{}
	r.Arguments("1", "2", "3")
	r.Arguments("4")
	r.Option("stringkey", "stringvalue")
	r.Option("bytekey", []byte("bytevalue"))
	r.Option("boolkey", true)
	r.Option("otherkey", now)
	r.Header("some-header", "header-value")
	r.Header("some-header-2", "header-value-2")

	is.Equal(r.args, []string{"1", "2", "3", "4"})
	is.Equal(r.opts, map[string]string{
		"stringkey": "stringvalue",
		"bytekey":   "bytevalue",
		"boolkey":   "true",
		"otherkey":  now.String(),
	})
	is.Equal(r.headers, map[string]string{
		"some-header":   "header-value",
		"some-header-2": "header-value-2",
	})
}
