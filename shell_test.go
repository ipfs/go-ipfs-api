package shell

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/TRON-US/go-btfs-api/options"
	"github.com/tron-us/go-btfs-common/crypto"

	"github.com/cheekybits/is"
	"github.com/ipfs/go-cid"
	"github.com/jarcoal/httpmock"
	mh "github.com/multiformats/go-multihash"
	"github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
)

const (
	examplesHash = "QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg2G"
	shellUrl     = "localhost:5001"
	privateKey   = `CAISIMozgNaJkUdEzxtIguLWpzjRxcEDwDDC0gNv13r8fa53`
)

func peerSessionSignature(privateKey string, timestamp string, peerId string) string {
	privKey, err := crypto.ToPrivKey(privateKey)
	if err != nil {
		log.Error(err)
	}
	timeNonce := fmt.Sprint("", peerId, timestamp)
	timeNonceBytes := []byte(timeNonce)
	sign, err := privKey.Sign(timeNonceBytes)
	if err != nil {
		log.Error(err)
	}
	return string(sign)
}

func upload(s *Shell) (string, string) {

	offlineTimeStamp := time.Stamp
	hostname, _ := os.Hostname()
	pid := string(os.Getpid())

	pref := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1, // default length
	}

	peerId, _ := pref.Sum([]byte(fmt.Sprint("", hostname, pid, offlineTimeStamp)))
	timeNonce := fmt.Sprint("", offlineTimeStamp, examplesHash, peerId)
	options := func(rb *RequestBuilder) error {
		rb.Option("peer-session-signature", "")
		return nil
	}

	sessionId, _ := s.StorageUpload(examplesHash, peerId.String(), timeNonce, peerSessionSignature(privateKey, offlineTimeStamp, peerId.String()), options)
	return sessionId, peerId.String()
}

func TestAdd(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	mhash, err := s.Add(bytes.NewBufferString("Hello IPFS Shell tests"))
	is.Nil(err)
	is.Equal(mhash, "QmUfZ9rAdhV5ioBzXKdUTh2ZNsz9bzbkaLVyQ8uc8pj21F")
}

func TestUpload(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	mhash, err := s.Add(bytes.NewBufferString("Hello IPFS Shell tests"))
	is.Nil(err)
	is.Equal(mhash, "QmUfZ9rAdhV5ioBzXKdUTh2ZNsz9bzbkaLVyQ8uc8pj21F")
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
		topic = "test"

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
	is.Nil(s.PubSubPublish(topic, "Hello World!"))
	t.Log("pub: done")

	t.Log("next()...")
	r, err := sub.Next()
	t.Log("next: done. ")

	is.Nil(err)
	is.NotNil(r)
	is.Equal(r.Data, "Hello World!")

	sub2, err := s.PubSubSubscribe(topic)
	is.Nil(err)
	is.NotNil(sub2)

	is.Nil(s.PubSubPublish(topic, "Hallo Welt!"))

	r, err = sub2.Next()
	is.Nil(err)
	is.NotNil(r)
	is.Equal(r.Data, "Hallo Welt!")

	r, err = sub.Next()
	is.NotNil(r)
	is.Nil(err)
	is.Equal(r.Data, "Hallo Welt!")

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

	c, err := s.DagPut(`{"x": "abc","y":"def"}`, "json", "cbor")
	is.Nil(err)
	is.Equal(c, "zdpuAt47YjE9XTgSxUBkiYCbmnktKajQNheQBGASHj3FfYf8M")
}

func TestDagPutWithOpts(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	c, err := s.DagPutWithOpts(`{"x": "abc","y":"def"}`, options.Dag.Pin("true"))
	is.Nil(err)
	is.Equal(c, "zdpuAt47YjE9XTgSxUBkiYCbmnktKajQNheQBGASHj3FfYf8M")
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

type ShardItem struct {
	ContractId string `json:contractid`
	Price      string `json:price`
	Host       string `json:host`
	Status     string `json:status`
}
type StatusResponse struct {
	Status   string      `json:staus`
	Filehash string      `json:filehash`
	Shards   []ShardItem `json:shards`
}

type ContractBatchResponse struct {
	Contracts [30]ContractItem
}

type StatusRequest struct {
	SessionId string `json:sessionid`
}
type ContractBatchRequest struct {
	SessionId              string `json:sessionid`
	PeerId                 string `json:peerid`
	NonceTimeStamp         int64  `json:noncetimestamp`
	UploadSessionSignature string `json:uploadsessionsignature`
	SessionStatus          string `json:sessionstatus`
}
type AccessStatus interface {
	getStatus() (StatusResponse, error)
}
type Status struct {
	Request StatusRequest
}

type ContractBatch struct {
	Request ContractBatchRequest
}

func (cb ContractBatch) getContractsBatch() (ContractBatchResponse, error) {
	// Access the API
	url := fmt.Sprintf("http://localhost:5001/api/v1/storage/upload/getcontractbatch?session-id=%s&peer-id=%s&nonce-timestamp=%v&upload-session-signature=%s&session-status=%s", cb.Request.SessionId, cb.Request.PeerId, cb.Request.NonceTimeStamp, cb.Request.UploadSessionSignature, cb.Request.SessionStatus)

	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err)
		return ContractBatchResponse{}, err
	}

	// Create an HTTP Client
	client := &http.Client{}

	// Send the request via a client
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return ContractBatchResponse{}, err
	}

	// Defer the closing of the body
	defer resp.Body.Close()

	var rqResp ContractBatchResponse
	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&rqResp); err != nil {
		log.Error(err)
		return ContractBatchResponse{}, err
	}

	return rqResp, nil
}
func (s Status) getStatus() (StatusResponse, error) {

	// Access the API
	url := fmt.Sprintf("http://localhost:5001/api/v1/storage/upload/status?session-id=%s", s.Request.SessionId)

	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err)
		return StatusResponse{}, err
	}

	// Create an HTTP Client
	client := &http.Client{}

	// Send the request via a client
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return StatusResponse{}, err
	}

	// Defer the closing of the body
	defer resp.Body.Close()

	var rqResp StatusResponse
	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&rqResp); err != nil {
		log.Error(err)
		return StatusResponse{}, err
	}

	return rqResp, nil
}
func TestStorageUploadInitStatus(t *testing.T) {

	assert := assert.New(t)

	//activate http mock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	//create a responder to cycle through the offline signing states
	responder := httpmock.NewStringResponder(200, `{
	  "Status": "uninitialized",
	  "FileHash": "QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg15"
	}`)
	responder2 := httpmock.NewStringResponder(200, `{
	  "Status": "initSignReadyForEscrow",
	  "FileHash": "QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg15"
	}`)
	responder3 := httpmock.NewStringResponder(200, `{
	  "Contracts": [
		{
		  "Contract":"QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg15",
		  "ContractId": 1
		},
		{
		  "Contract":"QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg16",
		  "ContractId": 2
		},
		{
		  "Contract":"QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg17",
		  "ContractId": 3
		}
	  ]
	}`)
	responder4 := httpmock.NewStringResponder(200, `{
	  "Status": "initSignReadyForGuard",
	  "FileHash": "QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg15"
	}`)
	responder5 := httpmock.NewStringResponder(200, `{
	  "Contracts": [
		{
		  "Contract":"QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg18",
		  "ContractId": 1
		},
		{
		  "Contract":"QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg19",
		  "ContractId": 2
		},
		{
		  "Contract":"QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg20",
		  "ContractId": 3
		}
	  ]
	}`)

	//create a test
	status := Status{Request: StatusRequest{"1"}}

	escrowContract := ContractBatch{Request: ContractBatchRequest{"1", "1", 1, "TTT"}}

	//create request string and register string with responder
	url := fmt.Sprintf("http://localhost:5001/api/v1/storage/upload/status?session-id=%s", status.Request.SessionId)
	httpmock.RegisterResponder("GET", url, responder)
	//call the endpoint to retreive the mocked values of the responder
	resp, _ := status.getStatus()
	assert.Equal(resp.Status, "uninitialized")
	httpmock.Reset()

	//wait for a few seconds

	url2 := fmt.Sprintf("http://localhost:5001/api/v1/storage/upload/status?session-id=%s", status.Request.SessionId)
	httpmock.RegisterResponder("GET", url2, responder2)
	resp2, _ := status.getStatus()
	assert.Equal(resp2.Status, "initSignReadyForEscrow")
	httpmock.Reset()

	//get contracts
	url3 := fmt.Sprintf("http://localhost:5001/api/v1/storage/upload/getcontractbatch?session-id=%s&peer-id=%s&nonce-timestamp=%v&upload-session-signature=%s&session-status=&%s", escrowContract.Request.SessionId, escrowContract.Request.PeerId, escrowContract.Request.NonceTimeStamp, escrowContract.Request.UploadSessionSignature, resp2)
	httpmock.RegisterResponder("GET", url3, responder3)
	resp3, _ := escrowContract.getContractsBatch()
	assert.Equal(resp3.Contracts[0].Contract, "QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg15")
	httpmock.Reset()

	//sign and send contracts

	url4 := fmt.Sprintf("http://localhost:5001/api/v1/storage/upload/status?session-id=%s", status.Request.SessionId)
	httpmock.RegisterResponder("GET", url4, responder4)
	resp4, _ := status.getStatus()
	assert.Equal(resp4.Status, "initSignReadyForGuard")
	httpmock.Reset()

	//get contracts

	url5 := fmt.Sprintf("http://localhost:5001/api/v1/storage/upload/getcontractbatch?session-id=%s&peer-id=%s&nonce-timestamp=%v&upload-session-signature=%s", escrowContract.Request.SessionId, escrowContract.Request.PeerId, escrowContract.Request.NonceTimeStamp, escrowContract.Request.UploadSessionSignature)
	httpmock.RegisterResponder("GET", url5, responder5)
	resp5, _ := escrowContract.getContractsBatch()
	assert.Equal(resp5.Contracts[0].Contract, "QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg15")
	httpmock.Reset()

	//sign and send contracts

}

func TestStorageUploadPaymentStatus(t *testing.T) {

	assert := assert.New(t)

	//activate http mock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	//create a responder to cycle through the offline signing states
	responder := httpmock.NewStringResponder(200, `{
	  "Status": "uninitialized",
	  "FileHash": "QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg15"
	}`)
	responder2 := httpmock.NewStringResponder(200, `{
	  "Status": "initSignReadyForEscrow",
	  "FileHash": "QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg15"
	}`)
	responder3 := httpmock.NewStringResponder(200, `{
	  "Contracts": [
		{
		  "Contract":"QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg15",
		  "ContractId": 1
		},
		{
		  "Contract":"QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg16",
		  "ContractId": 2
		},
		{
		  "Contract":"QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg17",
		  "ContractId": 3
		}
	  ]
	}`)
	responder4 := httpmock.NewStringResponder(200, `{
	  "Status": "initSignReadyForGuard",
	  "FileHash": "QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg15"
	}`)
	responder5 := httpmock.NewStringResponder(200, `{
	  "Contracts": [
		{
		  "Contract":"QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg18",
		  "ContractId": 1
		},
		{
		  "Contract":"QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg19",
		  "ContractId": 2
		},
		{
		  "Contract":"QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg20",
		  "ContractId": 3
		}
	  ]
	}`)

	//create a test
	status := Status{Request: StatusRequest{"1"}}

	escrowContract := ContractBatch{Request: ContractBatchRequest{"1", "1", 1, "TTT", "XXX"}}

	//create request string and register string with responder
	url := fmt.Sprintf("http://localhost:5001/api/v1/storage/upload/status?session-id=%s", status.Request.SessionId)
	httpmock.RegisterResponder("GET", url, responder)
	//call the endpoint to retreive the mocked values of the responder
	resp, _ := status.getStatus()
	assert.Equal(resp.Status, "uninitialized")
	httpmock.Reset()

	//wait for a few seconds

	url2 := fmt.Sprintf("http://localhost:5001/api/v1/storage/upload/status?session-id=%s", status.Request.SessionId)
	httpmock.RegisterResponder("GET", url2, responder2)
	resp2, _ := status.getStatus()
	assert.Equal(resp2.Status, "initSignReadyForEscrow")
	httpmock.Reset()

	//get contracts
	url3 := fmt.Sprintf("http://localhost:5001/api/v1/storage/upload/getcontractbatch?session-id=%s&peer-id=%s&nonce-timestamp=%v&upload-session-signature=%s", escrowContract.Request.SessionId, escrowContract.Request.PeerId, escrowContract.Request.NonceTimeStamp, escrowContract.Request.UploadSessionSignature)
	httpmock.RegisterResponder("GET", url3, responder3)
	resp3, _ := escrowContract.getContractsBatch()
	assert.Equal(resp3.Contracts[0].Contract, "QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg15")
	httpmock.Reset()

	//sign and send contracts

	url4 := fmt.Sprintf("http://localhost:5001/api/v1/storage/upload/status?session-id=%s", status.Request.SessionId)
	httpmock.RegisterResponder("GET", url4, responder4)
	resp4, _ := status.getStatus()
	assert.Equal(resp4.Status, "initSignReadyForGuard")
	httpmock.Reset()

	//get contracts

	url5 := fmt.Sprintf("http://localhost:5001/api/v1/storage/upload/getcontractbatch?session-id=%s&peer-id=%s&nonce-timestamp=%v&upload-session-signature=%s", escrowContract.Request.SessionId, escrowContract.Request.PeerId, escrowContract.Request.NonceTimeStamp, escrowContract.Request.UploadSessionSignature)
	httpmock.RegisterResponder("GET", url5, responder5)
	resp5, _ := escrowContract.getContractsBatch()
	assert.Equal(resp5.Contracts[0].Contract, "QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg15")
	httpmock.Reset()

	//sign and send contracts

}

func TestStorageUpload(t *testing.T) {
	is := is.New(t)
	s := NewShell("https://storageupload.free.beeceptor.com")
	offlineTimeStamp := time.Stamp
	hostname, err := os.Hostname()
	pid := string(os.Getpid())

	pref := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1, // default length
	}

	peerId, err := pref.Sum([]byte(fmt.Sprint("", hostname, pid, offlineTimeStamp)))
	timeNonce := fmt.Sprint("", offlineTimeStamp, examplesHash, peerId)
	options := func(rb *RequestBuilder) error {
		return nil
	}

	sessionId, err := s.StorageUpload(examplesHash, peerId.String(), timeNonce, peerSessionSignature(privateKey, offlineTimeStamp, peerId.String()), options)
	is.Nil(err)
	is.NotNil(sessionId)
}

/*func TestStorageUploadGetContractBatch(t *testing.T) {

	//check status for initSignReadyForEscrow:
	//get ContractBatch

	//check status for initSignReadyForGuard:
	//get ContractBatch

	type Contracts struct {
		Contracts [30]ContractItem
	}

	var contracts Contracts
	errEscrow := json.Unmarshal(contractsEscrow, &contracts)
	if errEscrow != nil {
		log.Error(errEscrow)
	}

	errGuard := json.Unmarshal(contractsGuard, &contracts)
	if errGuard != nil {
		log.Error(errGuard)
	}

	fmt.Println("Here is the first contract: ", contracts.Contracts[0].Contract)

	is := is.New(t)
	s := NewShell(shellUrl)
	sessionId, peerId := upload(s)
	offlineTimeStamp := time.Stamp
	storage, err := s.StorageUploadStatus(sessionId, peerId, offlineTimeStamp, peerSessionSignature(privateKey, offlineTimeStamp, peerId))
	is.Equal(storage.Status, "init-sign")
	respContracts, err := s.StorageUploadGetContractBatch(sessionId, peerId, offlineTimeStamp, peerSessionSignature(privateKey, offlineTimeStamp, peerId))
	is.Nil(err)
	is.Equal(respContracts, raw)
}*/
/*
func TestShell_StorageUploadSignBatch(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)
	sessionId, peerId := upload(s)
	offlineTimeStamp := time.Stamp
	privateKey := "f65a52c238ffb69b93d954c769ea992d"
	storage, err := s.StorageUploadStatus(sessionId, peerId, offlineTimeStamp, peerSessionSignature(privateKey, offlineTimeStamp, peerId))
	is.Equal(storage.Status, "init-sign")
	respContracts, err := s.StorageUploadGetContractBatch(sessionId, peerId, offlineTimeStamp, peerSessionSignature(privateKey, offlineTimeStamp, peerId))
	respContractBytes, err := s.StorageUploadSignBatch(sessionId, peerId, offlineTimeStamp, respContracts , peerSessionSignature(privateKey, offlineTimeStamp, peerId))
	is.Nil(err)
	is.Equal(respContractBytes, Sign(privateKey, respContracts))
}
*/
/*
func TestStorageUploadPaymentStatus(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)
	sessionId, peerId := upload(s)
	offlineTimeStamp := time.Stamp

	initStatus, err := s.StorageUploadStatus(sessionId, peerId, offlineTimeStamp, peerSessionSignature(privateKey, offlineTimeStamp, peerId))
	is.Equal(initStatus, "init-sign")

	//
	respUnsignedContracts, err := s.StorageUploadGetContractBatch(sessionId, peerId, offlineTimeStamp, peerSessionSignature(privateKey, offlineTimeStamp, peerId))
	respSignedInit, err := s.StorageUploadSignBatch(sessionId, peerId, offlineNonceTimeStamp, respUnsignedContracts )
	paymentStatus, err := s.StorageUploadStatus(sessionId, peerId, offlineNonceTimeStamp)
	is.Equal(paymentStatus, "payment-sign")
	is.Nil(err)
	is.Equal(respSignedInit, "okay")

    //
	respUnsignedPaymentContracts, err := s.StorageUploadGetContractBatch(sessionId, peerId, offlineNonceTimeStamp)
	respSignedPayment, err := s.StorageUploadSignBatch(sessionId, peerId, offlineNonceTimeStamp, respUnsignedPaymentContracts )
	storage, err := s.StorageUploadStatus(sessionId, peerId, offlineNonceTimeStamp)
	is.Equal(storage.Status, "complete")
	is.Nil(err)
	is.Equal(respSignedPayment, "okay")
}*/
