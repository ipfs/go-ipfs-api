package shell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/cheekybits/is"
)

func TestMain(m *testing.M) {
	s := NewShell(shellUrl)
	files := []string{"about", "readme"}
	for _, f := range files {
		file, err := os.Open(fmt.Sprintf("./testdata/%s", f))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open test data file: %v\n", err)
			os.Exit(1)
		}

		err = s.FilesWrite(context.Background(), fmt.Sprintf("/testdata/%s", f), file, FilesWrite.Parents(true), FilesWrite.Create(true))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write test data, not running tests: %v\n", err)
			os.Exit(1)
		}
	}

	stat, err := s.FilesStat(context.Background(), "/testdata")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stat test data, not running tests: %v\n", err)
		os.Exit(1)
	}

	expectedTestdataCIDString := "QmfZtacPc5nch976ZsiBw6nhLmTzy5JjW2pzZg8j7GjqWq"
	if stat.Hash != expectedTestdataCIDString {
		fmt.Fprintf(os.Stderr, "CID of /testdata is %s which does not match the expected %s, not running tests\n", stat.Hash, expectedTestdataCIDString)
		os.Exit(1)
	}

	exitVal := m.Run()
	if err := s.FilesRm(context.Background(), "/testdata", true); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to remove test data: %v\n", err)
		os.Exit(1)
	}
	os.Exit(exitVal)
}

func TestFilesChcid(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	err := s.FilesChcid(context.Background(), "/testdata", FilesChcid.Hash("sha3-256"))
	is.Nil(err)

	stat, err := s.FilesStat(context.Background(), "/testdata")
	is.Nil(err)
	is.Equal(stat.Hash, "bafybmigo44bvq5f4u2oswr7cilvlilftjekr4iilwxuxjj326hchztmk2m")

	err = s.FilesChcid(context.Background(), "/testdata", FilesChcid.CidVersion(0))
	is.Nil(err)

	stat, err = s.FilesStat(context.Background(), "/testdata")
	is.Nil(err)
	is.Equal(stat.Hash, "QmfZtacPc5nch976ZsiBw6nhLmTzy5JjW2pzZg8j7GjqWq")
}

func TestFilesCp(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	err := s.FilesCp(context.Background(), "/testdata/readme", "/testdata/readme2")
	is.Nil(err)

	stat, err := s.FilesStat(context.Background(), "/testdata/readme2")
	is.Nil(err)
	is.Equal(stat.Hash, "QmfZt7xPekp7npSM6DHDUnFseAiNZQs7wq6muH9o99RsCB")

	err = s.FilesRm(context.Background(), "/testdata/readme2", true)
	is.Nil(err)
}

func TestFilesCParents(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	err := s.FilesCp(context.Background(), "/testdata/readme", "/dirs/should/be/created/readme", FilesCp.Parents(true))
	is.Nil(err)

	stat, err := s.FilesStat(context.Background(), "/dirs/should/be/created/readme")
	is.Nil(err)
	is.Equal(stat.Hash, "QmfZt7xPekp7npSM6DHDUnFseAiNZQs7wq6muH9o99RsCB")

	err = s.FilesRm(context.Background(), "/dirs/should/be/created/readme", true)
	is.Nil(err)
}

func TestFilesFlush(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	cid, err := s.FilesFlush(context.Background(), "/testdata")
	is.Nil(err)
	is.Equal(cid, "QmfZtacPc5nch976ZsiBw6nhLmTzy5JjW2pzZg8j7GjqWq")
}

func TestFilesLs(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	list, err := s.FilesLs(context.Background(), "/testdata", FilesLs.Stat(true))
	is.Nil(err)

	is.Equal(len(list), 2)
	is.Equal(list[0].Name, "about")
	is.Equal(list[0].Hash, "QmUdTMirfB6dXArAZuB8V2QGcSVvRDZNa1XCquG6xR2cGK")
	is.Equal(list[1].Name, "readme")
	is.Equal(list[1].Hash, "QmfZt7xPekp7npSM6DHDUnFseAiNZQs7wq6muH9o99RsCB")
}

func TestFilesMkdir(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	err := s.FilesMkdir(context.Background(), "/testdata/dir1/dir2", FilesMkdir.Parents(true), FilesMkdir.CidVersion(1), FilesMkdir.Hash("sha3-256"))
	is.Nil(err)

	err = s.FilesMkdir(context.Background(), "/testdata/dir3/dir4")
	is.NotNil(err)

	err = s.FilesRm(context.Background(), "/testdata/dir1", true)
	is.Nil(err)
}

func TestFilesMv(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	err := s.FilesMv(context.Background(), "/testdata/readme", "/testdata/readme2")
	is.Nil(err)

	stat, err := s.FilesStat(context.Background(), "/testdata/readme2")
	is.Nil(err)
	is.Equal(stat.Hash, "QmfZt7xPekp7npSM6DHDUnFseAiNZQs7wq6muH9o99RsCB")

	_, err = s.FilesStat(context.Background(), "/testdata/readme")
	is.NotNil(err)

	err = s.FilesMv(context.Background(), "/testdata/readme2", "/testdata/readme")
	is.Nil(err)
}

func TestFilesRead(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	reader, err := s.FilesRead(context.Background(), "/testdata/readme", FilesRead.Offset(0), FilesRead.Count(5))
	is.Nil(err)

	resBytes, err := io.ReadAll(reader)
	is.Nil(err)
	is.Equal(string(resBytes), "Hello")
}

func TestFilesRm(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	file, _ := os.ReadFile("./testdata/ping")
	err := s.FilesWrite(context.Background(), "/testdata/dir1/ping", bytes.NewBuffer(file), FilesWrite.Parents(true), FilesWrite.Create(true))
	is.Nil(err)

	err = s.FilesRm(context.Background(), "/testdata/dir1", false)
	is.NotNil(err)

	err = s.FilesRm(context.Background(), "/testdata/dir1", true)
	is.Nil(err)
}

func TestFilesStat(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	res, err := s.FilesStat(context.Background(), "/testdata")
	is.Nil(err)
	is.Equal(res.Hash, "QmfZtacPc5nch976ZsiBw6nhLmTzy5JjW2pzZg8j7GjqWq")
	is.Equal(res.Size, 0)
	is.Equal(res.Type, "directory")

	res, err = s.FilesStat(context.Background(), "/testdata", FilesStat.WithLocal(true))
	is.Nil(err)
	is.Equal(res.WithLocality, true)
	is.Equal(res.Local, true)
	is.Equal(res.SizeLocal, 2997)
}

func TestFilesWrite(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	file, err := os.ReadFile("./testdata/ping")
	is.Nil(err)

	err = s.FilesWrite(context.Background(), "/testdata/ping", bytes.NewBuffer(file), FilesWrite.Create(true), FilesWrite.RawLeaves(true), FilesWrite.CidVersion(1), FilesWrite.Hash("sha3-256"))
	is.Nil(err)

	reader, err := s.FilesRead(context.Background(), "/testdata/ping")
	is.Nil(err)

	resBytes, err := io.ReadAll(reader)
	is.Nil(err)
	is.Equal(string(resBytes), "ipfs")

	file, err = os.ReadFile("./testdata/ping")
	is.Nil(err)

	err = s.FilesWrite(context.Background(), "/testdata/ping", bytes.NewBuffer(file), FilesWrite.Offset(0), FilesWrite.Count(2), FilesWrite.Truncate(true))
	is.Nil(err)

	reader, err = s.FilesRead(context.Background(), "/testdata/ping")
	is.Nil(err)

	resBytes, err = io.ReadAll(reader)
	is.Nil(err)
	is.Equal(string(resBytes), "ip")

	err = s.FilesRm(context.Background(), "/testdata/ping", true)
	is.Nil(err)
}
