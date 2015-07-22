package main

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/whyrusleeping/ipfs-shell"

	u "github.com/ipfs/go-ipfs/util"
)

var sh *shell.Shell

func randString() string {
	alpha := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	l := rand.Intn(10) + 2

	var s string
	for i := 0; i < l; i++ {
		s += string([]byte{alpha[rand.Intn(len(alpha))]})
	}
	return s
}

func makeRandomObject() (string, error) {
	// do some math to make a size
	x := rand.Intn(120) + 1
	y := rand.Intn(120) + 1
	z := rand.Intn(120) + 1
	size := x * y * z

	r := io.LimitReader(u.NewTimeSeededRand(), int64(size))
	return sh.Add(r)
}

func makeRandomDir(depth int) (string, error) {
	if depth <= 0 {
		return makeRandomObject()
	}
	empty, err := sh.NewObject("unixfs-dir")
	if err != nil {
		return "", err
	}

	curdir := empty
	for i := 0; i < rand.Intn(8)+2; i++ {
		var obj string
		if rand.Intn(2) == 1 {
			obj, err = makeRandomObject()
			if err != nil {
				return "", err
			}
		} else {
			obj, err = makeRandomDir(depth - 1)
			if err != nil {
				return "", err
			}
		}

		name := randString()
		nobj, err := sh.PatchLink(curdir, name, obj, true)
		if err != nil {
			return "", err
		}
		curdir = nobj
	}

	return curdir, nil
}

func main() {
	sh = shell.NewShell("localhost:5001")
	out, err := makeRandomDir(10)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(out)
}
