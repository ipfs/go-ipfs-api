package shell

import (
	gohttp "net/http"

	"github.com/ipfs/go-ipfs-api/legacy"
)

func NewLocalShell() *legacy.Shell {
	return legacy.NewLocalShell()
}

func NewShell(url string) *legacy.Shell {
	return legacy.NewShell(url)
}

func NewShellWithClient(url string, c *gohttp.Client) *legacy.Shell {
	return legacy.NewShellWithClient(url, c)
}
