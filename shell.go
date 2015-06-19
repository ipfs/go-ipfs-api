// package shell implements a remote API interface for a running ipfs daemon
package shell

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	cmds "github.com/ipfs/go-ipfs/commands"
	files "github.com/ipfs/go-ipfs/commands/files"
	http "github.com/ipfs/go-ipfs/commands/http"
	cc "github.com/ipfs/go-ipfs/core/commands"

	tar "github.com/ipfs/go-ipfs/thirdparty/tar"
)

type Shell struct {
	client http.Client
}

func NewShell(url string) *Shell {
	return &Shell{http.NewClient(url)}
}

type IdOutput struct {
	ID              string
	PublicKey       string
	Addresses       []string
	AgentVersion    string
	ProtocolVersion string
}

// ID gets information about a given peer.  Arguments:
//
// peer: peer.ID of the node to look up.  If no peer is specified,
//   return information about the local peer.
func (s *Shell) ID(peer ...string) (*IdOutput, error) {
	if len(peer) > 1 {
		return nil, fmt.Errorf("Too many peer arguments")
	}
	ropts, err := cc.Root.GetOptions([]string{"id"})
	if err != nil {
		return nil, err
	}

	req, err := cmds.NewRequest([]string{"id"}, nil, peer, nil, cc.IDCmd, ropts)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Send(req)
	if err != nil {
		return nil, err
	}

	reader, err := resp.Reader()
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(reader)
	out := new(IdOutput)
	err = decoder.Decode(out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// Cat the content at the given path
func (s *Shell) Cat(path string) (io.Reader, error) {
	ropts, err := cc.Root.GetOptions([]string{"cat"})
	if err != nil {
		return nil, err
	}

	req, err := cmds.NewRequest([]string{"cat"}, nil, []string{path}, nil, cc.CatCmd, ropts)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Send(req)
	if err != nil {
		return nil, err
	}

	return resp.Reader()
}

// Add a file to ipfs from the given reader, returns the hash of the added file
func (s *Shell) Add(r io.Reader) (string, error) {
	ropts, err := cc.Root.GetOptions([]string{"add"})
	if err != nil {
		return "", err
	}

	slf := files.NewSliceFile("", []files.File{files.NewReaderFile("", ioutil.NopCloser(r), nil)})

	req, err := cmds.NewRequest([]string{"add"}, nil, nil, slf, cc.AddCmd, ropts)
	if err != nil {
		return "", err
	}

	resp, err := s.client.Send(req)
	if err != nil {
		return "", err
	}
	if resp.Error() != nil {
		return "", resp.Error()
	}

	read, err := resp.Reader()
	if err != nil {
		return "", err
	}

	dec := json.NewDecoder(read)
	out := struct{ Hash string }{}
	err = dec.Decode(&out)
	if err != nil {
		return "", err
	}

	return out.Hash, nil
}

// AddDir adds a directory recursively with all of the files under it
func (s *Shell) AddDir(dir string) (string, error) {
	ropts, err := cc.Root.GetOptions([]string{"add"})
	if err != nil {
		return "", err
	}

	dfi, err := os.Open(dir)
	if err != nil {
		return "", err
	}

	sf, err := files.NewSerialFile(dir, dfi)
	if err != nil {
		return "", err
	}
	slf := files.NewSliceFile(dir, []files.File{sf})

	req, err := cmds.NewRequest([]string{"add"}, nil, nil, slf, cc.AddCmd, ropts)
	if err != nil {
		return "", err
	}

	resp, err := s.client.Send(req)
	if err != nil {
		return "", err
	}
	if resp.Error() != nil {
		return "", resp.Error()
	}

	read, err := resp.Reader()
	if err != nil {
		return "", err
	}

	dec := json.NewDecoder(read)
	out := struct{ Hash string }{}
	var final string
	for {
		err = dec.Decode(&out)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		final = out.Hash
	}

	if final == "" {
		return "", errors.New("no results received")
	}

	return out.Hash, nil
}

// List entries at the given path
func (s *Shell) List(path string) ([]cc.Link, error) {
	ropts, err := cc.Root.GetOptions([]string{"ls"})
	if err != nil {
		return nil, err
	}

	req, err := cmds.NewRequest([]string{"ls"}, nil, []string{path}, nil, cc.LsCmd, ropts)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Send(req)
	if err != nil {
		return nil, err
	}

	if resp.Error() != nil {
		return nil, resp.Error()
	}

	read, err := resp.Reader()
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(read)
	out := struct{ Objects []cc.Object }{}
	err = dec.Decode(&out)
	if err != nil {
		return nil, err
	}

	return out.Objects[0].Links, nil
}

// Pin the given path
func (s *Shell) Pin(path string) error {
	ropts, err := cc.Root.GetOptions([]string{"pin", "add"})
	if err != nil {
		return err
	}
	pinadd := cc.PinCmd.Subcommands["add"]

	req, err := cmds.NewRequest([]string{"pin", "add"}, nil, []string{path}, nil, pinadd, ropts)
	if err != nil {
		return err
	}
	req.SetOption("r", true)

	resp, err := s.client.Send(req)
	if err != nil {
		return err
	}
	if resp.Error() != nil {
		return resp.Error()
	}

	read, err := resp.Reader()
	if err != nil {
		return err
	}

	out, err := ioutil.ReadAll(read)
	if err != nil {
		return err
	}

	fmt.Println(string(out))
	return nil
}

type PeerInfo struct {
	Addrs []string
	ID    string
}

func (s *Shell) FindPeer(peer string) (*PeerInfo, error) {
	ropts, err := cc.Root.GetOptions([]string{"dht", "findpeer"})
	if err != nil {
		return nil, err
	}
	fpeer := cc.DhtCmd.Subcommands["findpeer"]

	req, err := cmds.NewRequest([]string{"dht", "findpeer"}, nil, []string{peer}, nil, fpeer, ropts)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Send(req)
	if err != nil {
		return nil, err
	}
	if resp.Error() != nil {
		return nil, resp.Error()
	}

	read, err := resp.Reader()
	if err != nil {
		return nil, err
	}

	str := struct {
		Responses []PeerInfo
	}{}
	err = json.NewDecoder(read).Decode(&str)
	if err != nil {
		return nil, err
	}

	if len(str.Responses) == 0 {
		return nil, errors.New("peer not found")
	}

	return &str.Responses[0], nil
}

func (s *Shell) Refs(hash string, recursive bool) (<-chan string, error) {
	ropts, err := cc.Root.GetOptions([]string{"refs"})
	if err != nil {
		return nil, err
	}

	req, err := cmds.NewRequest([]string{"refs"}, nil, []string{hash}, nil, cc.RefsCmd, ropts)
	if err != nil {
		return nil, err
	}
	req.SetOption("r", recursive)

	resp, err := s.client.Send(req)
	if err != nil {
		return nil, err
	}
	if resp.Error() != nil {
		return nil, resp.Error()
	}

	read, err := resp.Reader()
	if err != nil {
		return nil, err
	}

	out := make(chan string)
	go func() {
		scan := bufio.NewScanner(read)
		for scan.Scan() {
			if len(scan.Text()) > 0 {
				out <- scan.Text()
			}
		}
		close(out)
	}()

	return out, nil
}

func (s *Shell) Patch(root, action string, args ...string) (string, error) {
	ropts, err := cc.Root.GetOptions([]string{"object", "patch"})
	if err != nil {
		return "", err
	}

	patchcmd := cc.ObjectCmd.Subcommand("patch")

	cmdargs := append([]string{root, action}, args...)
	req, err := cmds.NewRequest([]string{"object", "patch"}, nil, cmdargs, nil, patchcmd, ropts)
	if err != nil {
		return "", err
	}

	resp, err := s.client.Send(req)
	if err != nil {
		return "", err
	}
	if resp.Error() != nil {
		return "", resp.Error()
	}

	read, err := resp.Reader()
	if err != nil {
		return "", err
	}

	dec := json.NewDecoder(read)
	var out map[string]interface{}
	err = dec.Decode(&out)
	if err != nil {
		return "", err
	}

	hash, ok := out["Hash"]
	if !ok {
		return "", errors.New("no Hash field in command response")
	}

	return hash.(string), nil
}

func (s *Shell) Get(hash, outdir string) error {
	ropts, err := cc.Root.GetOptions([]string{"get"})
	if err != nil {
		return err
	}

	req, err := cmds.NewRequest([]string{"get"}, nil, []string{hash}, nil, cc.GetCmd, ropts)
	if err != nil {
		return err
	}

	resp, err := s.client.Send(req)
	if err != nil {
		return err
	}
	if resp.Error() != nil {
		return resp.Error()
	}

	read, err := resp.Reader()
	if err != nil {
		return err
	}

	extractor := &tar.Extractor{Path: outdir}
	return extractor.Extract(read)
}

func (s *Shell) NewObject(template string) (string, error) {
	ropts, err := cc.Root.GetOptions([]string{"object", "new"})
	if err != nil {
		return "", err
	}

	newcmd := cc.ObjectCmd.Subcommand("new")

	path := []string{"object", "new"}
	args := []string{}
	if template != "" {
		args = []string{template}
	}
	req, err := cmds.NewRequest(path, nil, args, nil, newcmd, ropts)
	if err != nil {
		return "", err
	}

	resp, err := s.client.Send(req)
	if err != nil {
		return "", err
	}
	if resp.Error() != nil {
		return "", resp.Error()
	}

	read, err := resp.Reader()
	if err != nil {
		return "", err
	}

	dec := json.NewDecoder(read)
	var out map[string]interface{}
	err = dec.Decode(&out)
	if err != nil {
		return "", err
	}

	hash, ok := out["Hash"]
	if !ok {
		return "", errors.New("no Hash field in command response")
	}

	return hash.(string), nil
}
