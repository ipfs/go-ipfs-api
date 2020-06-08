package shell

import (
	"context"
	"io"

	files "github.com/ipfs/go-ipfs-files"
)

type MfsLsEntry struct {
	Name string
	Type uint8
	Size uint64
	Hash string
}

type filesLsOutput struct {
	Entries []*MfsLsEntry
}

type filesFlushOutput struct {
	Cid string
}

type filesStatOutput struct {
	Blocks         int
	CumulativeSize uint64
	Hash           string
	Local          bool
	Size           uint64
	SizeLocal      uint64
	Type           string
	WithLocality   bool
}

type filesOpt func(*RequestBuilder) error
type filesLs struct{}
type filesChcid struct{}
type filesMkdir struct{}
type filesRead struct{}
type filesWrite struct{}
type filesStat struct{}

var (
	FilesLs    filesLs
	FilesChcid filesChcid
	FilesMkdir filesMkdir
	FilesRead  filesRead
	FilesWrite filesWrite
	FilesStat  filesStat
)

// Long use long listing format
func (filesLs) Long(long bool) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("long", long)
		return nil
	}
}

// U do not sort; list entries in directory order
func (filesLs) U(u bool) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("U", u)
		return nil
	}
}

// CidVersion cid version to use. (experimental)
func (filesChcid) CidVersion(version int) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("cid-version", version)
		return nil
	}
}

// Hash hash function to use. Will set Cid version to 1 if used. (experimental)
func (filesChcid) Hash(hash string) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("hash", hash)
		return nil
	}
}

// Parents no error if existing, make parent directories as needed
func (filesMkdir) Parents(parents bool) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("parents", parents)
		return nil
	}
}

// CidVersion cid version to use. (experimental)
func (filesMkdir) CidVersion(version int) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("cid-version", version)
		return nil
	}
}

// Hash hash function to use. Will set Cid version to 1 if used. (experimental)
func (filesMkdir) Hash(hash string) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("hash", hash)
		return nil
	}
}

// Offset byte offset to begin reading from
func (filesRead) Offset(offset int64) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("offset", offset)
		return nil
	}
}

// Count maximum number of bytes to read
func (filesRead) Count(count int64) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("count", count)
		return nil
	}
}

// Format print statistics in given format. Allowed tokens: <hash> <size> <cumulsize> <type> <childs>. Conflicts with other format options.
func (filesStat) Format(format string) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("format", format)
		return nil
	}
}

// Hash print only hash. Implies '--format=<hash>'. Conflicts with other format options.
func (filesStat) Hash(hash bool) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("hash", hash)
		return nil
	}
}

// Size print only size. Implies '--format=<cumulsize>'. Conflicts with other format options.
func (filesStat) Size(size bool) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("size", size)
		return nil
	}
}

// WithLocal compute the amount of the dag that is local, and if possible the total size.
func (filesStat) WithLocal(withLocal bool) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("with-local", withLocal)
		return nil
	}
}

// Offset byte offset to begin writing at
func (filesWrite) Offset(offset int64) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("offset", offset)
		return nil
	}
}

// Create create the file if it does not exist
func (filesWrite) Create(create bool) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("create", create)
		return nil
	}
}

// Parents make parent directories as needed
func (filesWrite) Parents(parents bool) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("parents", parents)
		return nil
	}
}

// Truncate truncate the file to size zero before writing
func (filesWrite) Truncate(truncate bool) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("truncate", truncate)
		return nil
	}
}

// Count maximum number of bytes to write
func (filesWrite) Count(count int64) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("count", count)
		return nil
	}
}

// RawLeaves use raw blocks for newly created leaf nodes. (experimental)
func (filesWrite) RawLeaves(rawLeaves bool) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("raw-leaves", rawLeaves)
		return nil
	}
}

// CidVersion cid version to use. (experimental)
func (filesWrite) CidVersion(version int) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("cid-version", version)
		return nil
	}
}

// Hash hash function to use. Will set Cid version to 1 if used. (experimental)
func (filesWrite) Hash(hash string) filesOpt {
	return func(rb *RequestBuilder) error {
		rb.Option("hash", hash)
		return nil
	}
}

// FilesChcid change the cid version or hash function of the root node of a given path
func (s *Shell) FilesChcid(path string, options ...filesOpt) error {
	if len(path) == 0 {
		path = "/"
	}

	rb := s.Request("files/chcid", path)
	for _, opt := range options {
		if err := opt(rb); err != nil {
			return err
		}
	}

	resp, err := rb.Send(context.Background())
	return handleResponse(resp, err)
}

// FilesCp copy any IPFS files and directories into MFS (or copy within MFS)
func (s *Shell) FilesCp(src string, dest string) error {
	resp, err := s.Request("files/cp", src, dest).
		Send(context.Background())
	return handleResponse(resp, err)
}

// FilesFlush flush a given path's data to disk
func (s *Shell) FilesFlush(path string) (string, error) {
	if len(path) == 0 {
		path = "/"
	}
	out := &filesFlushOutput{}
	if err := s.Request("files/flush", path).
		Exec(context.Background(), out); err != nil {
		return "", err
	}

	return out.Cid, nil
}

// FilesLs list directories in the local mutable namespace
func (s *Shell) FilesLs(path string, options ...filesOpt) ([]*MfsLsEntry, error) {
	if len(path) == 0 {
		path = "/"
	}

	var out filesLsOutput
	rb := s.Request("files/ls", path)
	for _, opt := range options {
		if err := opt(rb); err != nil {
			return nil, err
		}
	}
	if err := rb.Exec(context.Background(), &out); err != nil {
		return nil, err
	}
	return out.Entries, nil
}

// FilesMkdir make directories
func (s *Shell) FilesMkdir(path string, options ...filesOpt) error {
	rb := s.Request("files/mkdir", path)
	for _, opt := range options {
		if err := opt(rb); err != nil {
			return err
		}
	}

	resp, err := rb.Send(context.Background())
	return handleResponse(resp, err)
}

// FilesMv move files
func (s *Shell) FilesMv(src string, dest string) error {
	resp, err := s.Request("files/mv", src, dest).
		Send(context.Background())
	return handleResponse(resp, err)
}

// FilesRead read a file in a given MFS
func (s *Shell) FilesRead(path string, options ...filesOpt) (io.ReadCloser, error) {
	rb := s.Request("files/read", path)
	for _, opt := range options {
		if err := opt(rb); err != nil {
			return nil, err
		}
	}

	resp, err := rb.Send(context.Background())
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}

	return resp.Output, nil
}

// FilesRm remove a file
func (s *Shell) FilesRm(path string, force bool) error {
	resp, err := s.Request("files/rm", path).
		Option("force", force).
		Send(context.Background())
	return handleResponse(resp, err)
}

// FilesStat display file status
func (s *Shell) FilesStat(path string, options ...filesOpt) (*filesStatOutput, error) {
	out := &filesStatOutput{}

	rb := s.Request("files/stat", path)
	for _, opt := range options {
		if err := opt(rb); err != nil {
			return nil, err
		}
	}

	if err := rb.Exec(context.Background(), out); err != nil {
		return nil, err
	}

	return out, nil
}

// FilesWrite write to a mutable file in a given filesystem
func (s *Shell) FilesWrite(path string, data io.Reader, options ...filesOpt) error {
	fr := files.NewReaderFile(data)
	slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
	fileReader := files.NewMultiFileReader(slf, true)

	rb := s.Request("files/write", path)
	for _, opt := range options {
		if err := opt(rb); err != nil {
			return err
		}
	}

	resp, err := rb.Body(fileReader).Send(context.Background())
	return handleResponse(resp, err)
}

func handleResponse(resp *Response, err error) error {
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return resp.Error
	}

	return nil
}
