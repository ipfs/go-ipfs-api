package shell

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cheekybits/is"
	files "github.com/ipfs/boxo/files"

	"github.com/ipfs/go-ipfs-api/options"
)

const (
	examplesHash = "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"
	shellUrl     = "localhost:5001"
)

func TestAdd(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	mhash, err := s.Add(bytes.NewBufferString("Hello IPFS Shell tests"))
	is.Nil(err)
	is.Equal(mhash, "QmUfZ9rAdhV5ioBzXKdUTh2ZNsz9bzbkaLVyQ8uc8pj21F")

	mhash, err = s.Add(bytes.NewBufferString("Hello IPFS Shell tests"), Hash("sha3-256"))
	is.Nil(err)
	is.Equal(mhash, "bafkrmidz7cuqruceo2hocadpdjppcsi7qw6dypz3jhsae2qda6sexdk6z4")

	mhash, err = s.Add(bytes.NewBufferString("Hello IPFS Shell tests"), CidVersion(1))
	is.Nil(err)
	is.Equal(mhash, "bafkreia5cxdsptovvt7qykfcrg4xlpaerart45pfn5di4rbivunybstmii")
}

func TestRedirect(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	err := s.
		Request("/version").
		Exec(context.Background(), nil)
	is.NotNil(err)
	is.True(strings.Contains(err.Error(), "unexpected redirect"))
}

func TestAddWithCat(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)
	s.SetTimeout(1 * time.Second)

	rand := randString(32)

	mhash, err := s.Add(bytes.NewBufferString(rand))
	is.Nil(err)

	reader, err := s.Cat(mhash)
	is.Nil(err)

	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	catRand := buf.String()

	is.Equal(rand, catRand)
}

func TestAddOnlyHash(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)
	s.SetTimeout(1 * time.Second)

	rand := randString(32)

	mhash, err := s.Add(bytes.NewBufferString(rand), OnlyHash(true))
	is.Nil(err)

	_, err = s.Cat(mhash)
	is.Err(err) // we expect an http timeout error because `cat` won't find the `rand` string
}

func TestAddNoPin(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	h, err := s.Add(bytes.NewBufferString(randString(32)), Pin(false))
	is.Nil(err)

	pins, err := s.Pins()
	is.Nil(err)

	_, ok := pins[h]
	is.False(ok)
}

func TestAddNoPinDeprecated(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	h, err := s.AddNoPin(bytes.NewBufferString(randString(32)))
	is.Nil(err)

	pins, err := s.Pins()
	is.Nil(err)

	_, ok := pins[h]
	is.False(ok)
}

func TestAddDir(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	cid, err := s.AddDir("./testdata")
	is.Nil(err)
	is.Equal(cid, "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv")
}

func TestAddDirWithCidV1(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	cid, err := s.AddDir("./testdata", CidVersion(1))
	is.Nil(err)
	is.Equal(cid, "bafybeibgegl5yqme2jehvvneapbq7he5ahi3tmk4cpmlagrggeji6hzayq")
}

func TestAddDirOffline(t *testing.T) {
	is := is.New(t)
	s := NewShell("0.0.0.0:1234") // connect to an invalid address

	_, err := s.AddDir("./testdata")
	is.Err(err)
}

func TestLocalShell(t *testing.T) {
	is := is.New(t)
	s := NewLocalShell()
	is.NotNil(s)

	mhash, err := s.Add(bytes.NewBufferString("Hello IPFS Shell tests"))
	is.Nil(err)
	is.Equal(mhash, "QmUfZ9rAdhV5ioBzXKdUTh2ZNsz9bzbkaLVyQ8uc8pj21F")
}

func TestCat(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	rc, err := s.Cat(fmt.Sprintf("/ipfs/%s/readme", examplesHash))
	is.Nil(err)

	md5 := md5.New()
	_, err = io.Copy(md5, rc)
	is.Nil(err)
	is.Equal(fmt.Sprintf("%x", md5.Sum(nil)), "3fdcaad186e79983a6920b4c7eeda949")
}

func TestList(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	list, err := s.List(fmt.Sprintf("/ipfs/%s", examplesHash))
	is.Nil(err)

	is.Equal(len(list), 7)

	// TODO: document difference in size between 'ipfs ls' and 'ipfs file ls -v'. additional object encoding in data block?
	expected := map[string]LsLink{
		"about":          {Type: TFile, Hash: "QmZTR5bcpQD7cFgTorqxZDYaew1Wqgfbd2ud9QqGPAkK2V", Name: "about", Size: 1677},
		"contact":        {Type: TFile, Hash: "QmYCvbfNbCwFR45HiNP45rwJgvatpiW38D961L5qAhUM5Y", Name: "contact", Size: 189},
		"help":           {Type: TFile, Hash: "QmY5heUM5qgRubMDD1og9fhCPA6QdkMp3QCwd4s7gJsyE7", Name: "help", Size: 311},
		"ping":           {Type: TFile, Hash: "QmejvEPop4D7YUadeGqYWmZxHhLc4JBUCzJJHWMzdcMe2y", Name: "ping", Size: 4},
		"quick-start":    {Type: TFile, Hash: "QmXgqKTbzdh83pQtKFb19SpMCpDDcKR2ujqk3pKph9aCNF", Name: "quick-start", Size: 1681},
		"readme":         {Type: TFile, Hash: "QmPZ9gcCEpqKTo6aq61g2nXGUhM4iCL3ewB6LDXZCtioEB", Name: "readme", Size: 1091},
		"security-notes": {Type: TFile, Hash: "QmQ5vhrL7uv6tuoN9KeVBwd4PwfQkXdVVmDLUZuTNxqgvm", Name: "security-notes", Size: 1162},
	}
	for _, l := range list {
		el, ok := expected[l.Name]
		is.True(ok)
		is.NotNil(el)
		is.Equal(*l, el)
	}
}

func TestFileList(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	list, err := s.FileList(fmt.Sprintf("/ipfs/%s", examplesHash))
	is.Nil(err)

	is.Equal(list.Type, "Directory")
	is.Equal(list.Size, 0)
	is.Equal(len(list.Links), 7)

	// TODO: document difference in sice betwen 'ipfs ls' and 'ipfs file ls -v'. additional object encoding in data block?
	expected := map[string]UnixLsLink{
		"about":          {Type: "File", Hash: "QmZTR5bcpQD7cFgTorqxZDYaew1Wqgfbd2ud9QqGPAkK2V", Name: "about", Size: 1677},
		"contact":        {Type: "File", Hash: "QmYCvbfNbCwFR45HiNP45rwJgvatpiW38D961L5qAhUM5Y", Name: "contact", Size: 189},
		"help":           {Type: "File", Hash: "QmY5heUM5qgRubMDD1og9fhCPA6QdkMp3QCwd4s7gJsyE7", Name: "help", Size: 311},
		"ping":           {Type: "File", Hash: "QmejvEPop4D7YUadeGqYWmZxHhLc4JBUCzJJHWMzdcMe2y", Name: "ping", Size: 4},
		"quick-start":    {Type: "File", Hash: "QmXgqKTbzdh83pQtKFb19SpMCpDDcKR2ujqk3pKph9aCNF", Name: "quick-start", Size: 1681},
		"readme":         {Type: "File", Hash: "QmPZ9gcCEpqKTo6aq61g2nXGUhM4iCL3ewB6LDXZCtioEB", Name: "readme", Size: 1091},
		"security-notes": {Type: "File", Hash: "QmQ5vhrL7uv6tuoN9KeVBwd4PwfQkXdVVmDLUZuTNxqgvm", Name: "security-notes", Size: 1162},
	}
	for _, l := range list.Links {
		el, ok := expected[l.Name]
		is.True(ok)
		is.NotNil(el)
		is.Equal(*l, el)
	}
}

func TestPins(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	// Add a thing, which pins it by default
	h, err := s.Add(bytes.NewBufferString("go-ipfs-api pins test 9F3D1F30-D12A-4024-9477-8F0C8E4B3A63"))
	is.Nil(err)

	pins, err := s.Pins()
	is.Nil(err)

	_, ok := pins[h]
	is.True(ok)

	err = s.Unpin(h)
	is.Nil(err)

	pins, err = s.Pins()
	is.Nil(err)

	_, ok = pins[h]
	is.False(ok)

	err = s.Pin(h)
	is.Nil(err)

	pins, err = s.Pins()
	is.Nil(err)

	info, ok := pins[h]
	is.True(ok)
	is.Equal(info.Type, RecursivePin)
}

func TestPinsOfType(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	// Add a thing, which pins it by default
	h, err := s.Add(bytes.NewBufferString("go-ipfs-api pins test 9F3D1F30-D12A-4024-9477-8F0C8E4B3A63"))
	is.Nil(err)

	pins, err := s.PinsOfType(context.Background(), RecursivePin)
	is.Nil(err)

	_, ok := pins[h]
	is.True(ok)

	err = s.Unpin(h)
	is.Nil(err)

	pins, err = s.Pins()
	is.Nil(err)

	_, ok = pins[h]
	is.False(ok)

	err = s.Pin(h)
	is.Nil(err)

	pins, err = s.Pins()
	is.Nil(err)

	info, ok := pins[h]
	is.True(ok)
	is.Equal(info.Type, RecursivePin)
}

func TestPinsStream(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	// Add a thing, which pins it by default
	h, err := s.Add(bytes.NewBufferString("go-ipfs-api pins test 0C7023F8-1FEC-4155-A8E0-432A5853F46B"))
	is.Nil(err)

	pinChan, err := s.PinsStream(context.Background())
	is.Nil(err)

	pins := accumulatePins(pinChan)

	_, ok := pins[h]
	is.True(ok)

	err = s.Unpin(h)
	is.Nil(err)

	pinChan, err = s.PinsStream(context.Background())
	is.Nil(err)

	pins = accumulatePins(pinChan)

	_, ok = pins[h]
	is.False(ok)

	err = s.Pin(h)
	is.Nil(err)

	pinChan, err = s.PinsStream(context.Background())
	is.Nil(err)

	pins = accumulatePins(pinChan)

	_type, ok := pins[h]
	is.True(ok)
	is.Equal(_type, RecursivePin)
}

func accumulatePins(pinChan <-chan PinStreamInfo) map[string]string {
	pins := make(map[string]string)
	for pin := range pinChan {
		pins[pin.Cid] = pin.Type
	}
	return pins
}

func TestPatch_rmLink(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)
	newRoot, err := s.Patch(examplesHash, "rm-link", "about")
	is.Nil(err)
	is.Equal(newRoot, "QmPmCJpciopaZnKcwymfQyRAEjXReR6UL2rdSfEscZfzcp")
}

func TestPatchLink(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)
	newRoot, err := s.PatchLink(examplesHash, "about", "QmUXTtySmd7LD4p6RG6rZW6RuUuPZXTtNMmRQ6DSQo3aMw", true)
	is.Nil(err)
	is.Equal(newRoot, "QmVfe7gesXf4t9JzWePqqib8QSifC1ypRBGeJHitSnF7fA")
	newRoot, err = s.PatchLink(examplesHash, "about", "QmUXTtySmd7LD4p6RG6rZW6RuUuPZXTtNMmRQ6DSQo3aMw", false)
	is.Nil(err)
	is.Equal(newRoot, "QmVfe7gesXf4t9JzWePqqib8QSifC1ypRBGeJHitSnF7fA")
	newHash, err := s.NewObject("unixfs-dir")
	is.Nil(err)
	_, err = s.PatchLink(newHash, "a/b/c", newHash, false)
	is.NotNil(err)
	newHash, err = s.PatchLink(newHash, "a/b/c", newHash, true)
	is.Nil(err)
	is.Equal(newHash, "QmQ5D3xbMWFQRC9BKqbvnSnHri31GqvtWG1G6rE8xAZf1J")
}

func TestResolvePath(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	childHash, err := s.ResolvePath(fmt.Sprintf("/ipfs/%s/about", examplesHash))
	is.Nil(err)
	is.Equal(childHash, "QmZTR5bcpQD7cFgTorqxZDYaew1Wqgfbd2ud9QqGPAkK2V")
}

func TestPubSub(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	var (
		topic    = "test\n topic\r\t with unsafe bytes"
		payload1 = "Hello\r\nWorld\t!"
		payload2 = "Hallo\r\nWelt\t!!11oneonę"

		sub *PubSubSubscription
		err error
	)

	t.Log("subscribing...")
	sub, err = s.PubSubSubscribe(topic)
	is.Nil(err)
	is.NotNil(sub)
	t.Log("sub: done")

	time.Sleep(10 * time.Millisecond)

	t.Log("publishing...")
	is.Nil(s.PubSubPublish(topic, payload1))
	t.Log("pub: done")

	t.Log("next()...")
	r, err := sub.Next()
	t.Log("next: done. ")

	is.Nil(err)
	is.NotNil(r)
	is.Equal(r.Data, payload1)

	sub2, err := s.PubSubSubscribe(topic)
	is.Nil(err)
	is.NotNil(sub2)

	is.Nil(s.PubSubPublish(topic, payload2))

	r, err = sub2.Next()
	is.Nil(err)
	is.NotNil(r)
	is.Equal(r.Data, payload2)

	r, err = sub.Next()
	is.NotNil(r)
	is.Nil(err)
	is.Equal(r.Data, payload2)

	is.Nil(sub.Cancel())
}

func TestObjectStat(t *testing.T) {
	obj := "QmZTR5bcpQD7cFgTorqxZDYaew1Wqgfbd2ud9QqGPAkK2V"
	is := is.New(t)
	s := NewShell(shellUrl)
	stat, err := s.ObjectStat("QmZTR5bcpQD7cFgTorqxZDYaew1Wqgfbd2ud9QqGPAkK2V")
	is.Nil(err)
	is.Equal(stat.Hash, obj)
	is.Equal(stat.LinksSize, 3)
	is.Equal(stat.CumulativeSize, 1688)
}

func TestDagPut(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	c, err := s.DagPut(`{"x": "abc","y":"def"}`, "dag-json", "dag-cbor")
	is.Nil(err)
	is.Equal(c, "bafyreidrm3r2k6vlxqp2fk47sboeycf7apddib47w7cyagrajtpaxxl2pi")
}

func TestDagPutWithOpts(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	c, err := s.DagPutWithOpts(`{"x": "abc","y":"def"}`, options.Dag.Pin("true"))
	is.Nil(err)
	is.Equal(c, "bafyreidrm3r2k6vlxqp2fk47sboeycf7apddib47w7cyagrajtpaxxl2pi")
}

func TestDagImport(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	carFile, err := os.ReadFile("./tests/test.car")
	is.Nil(err)

	_, err = s.DagImport(carFile, true, false)
	is.Nil(err)
}

func TestDagImportWithOpts(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	carFile, err := os.ReadFile("./tests/test.car")
	is.Nil(err)

	c, err := s.DagImportWithOpts(carFile, options.Dag.Stats(true), options.Dag.Silent(false))
	is.Nil(err)
	is.Equal(c.Roots[0].Root.Cid.Value, "bafybeibnhml2ecayjfa747ryfuy3ws5im6q4kscapqv7ajaspezwsw63ee")
	is.NotNil(c.Stats)
	is.Equal(c.Stats.BlockBytesCount, 173)
	is.Equal(c.Stats.BlockCount, 5)
}

func TestDagImportMultipleCARs(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	carFile1, err := os.ReadFile("./tests/cars/1.car")
	is.Nil(err)

	carFile2, err := os.ReadFile("./tests/cars/2.car")
	is.Nil(err)

	slf := files.NewSliceDirectory([]files.DirEntry{
		files.FileEntry("", files.NewReaderFile(bytes.NewReader(carFile1))),
		files.FileEntry("", files.NewReaderFile(bytes.NewReader(carFile2))),
	})

	dagImported, err := s.DagImportWithOpts(
		files.NewMultiFileReader(slf, true),
		options.Dag.Stats(true),
		options.Dag.Silent(false),
	)

	is.Nil(err)
	is.Equal(2, len(dagImported.Roots))
	is.Equal(411, dagImported.Stats.BlockBytesCount)
	is.Equal(11, dagImported.Stats.BlockCount)
}

func TestStatsBW(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)
	_, err := s.StatsBW(context.Background())
	is.Nil(err)
}

func TestSwarmPeers(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)
	_, err := s.SwarmPeers(context.Background())
	is.Nil(err)
}

// TestNewShellWithUnixSocket only check that http client is well configured to
// perform http request on unix socket address
func TestNewShellWithUnixSocket(t *testing.T) {
	is := is.New(t)

	// setup uds temporary dir
	path, err := os.MkdirTemp("", "uds-test")
	is.Nil(err)

	defer os.RemoveAll(path)

	// listen on sock path
	sockpath := filepath.Join(path, "sock")
	lsock, err := net.Listen("unix", sockpath)
	is.Nil(err)

	defer lsock.Close()

	// handle simple `hello` route
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v0/hello", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "Hello World\n")
	})

	go http.Serve(lsock, mux)

	// create shell with "/unix/<sockpath>" multiaddr
	shell := NewShell("/unix/" + sockpath)
	res, err := shell.Request("hello").Send(context.Background())
	is.Nil(err)

	defer res.Output.Close()

	// read hello world from body
	str, err := bufio.NewReader(res.Output).ReadString('\n')
	is.Nil(err)
	is.Equal(str, "Hello World\n")
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func TestRefs(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	cid, err := s.AddDir("./testdata")
	is.Nil(err)
	is.Equal(cid, "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv")
	refs, err := s.Refs(cid, false)
	is.Nil(err)
	expected := []string{
		"QmZTR5bcpQD7cFgTorqxZDYaew1Wqgfbd2ud9QqGPAkK2V",
		"QmYCvbfNbCwFR45HiNP45rwJgvatpiW38D961L5qAhUM5Y",
		"QmY5heUM5qgRubMDD1og9fhCPA6QdkMp3QCwd4s7gJsyE7",
		"QmejvEPop4D7YUadeGqYWmZxHhLc4JBUCzJJHWMzdcMe2y",
		"QmXgqKTbzdh83pQtKFb19SpMCpDDcKR2ujqk3pKph9aCNF",
		"QmPZ9gcCEpqKTo6aq61g2nXGUhM4iCL3ewB6LDXZCtioEB",
		"QmQ5vhrL7uv6tuoN9KeVBwd4PwfQkXdVVmDLUZuTNxqgvm",
	}
	var actual []string
	for r := range refs {
		actual = append(actual, r)
	}

	sort.Strings(expected)
	sort.Strings(actual)
	is.Equal(expected, actual)
}

func TestSwarmPeeringLs(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)
	_, err := s.SwarmPeeringLs(context.Background())
	is.Nil(err)
}

func TestSwarmPeeringAdd(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)
	addr := fmt.Sprintf("/ip4/10.10.10.10/tcp/4001/p2p/%s", examplesHash)
	_, err := s.SwarmPeeringAdd(context.Background(), addr)
	is.Nil(err)
}
