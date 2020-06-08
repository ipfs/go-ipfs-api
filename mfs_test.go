package shell

import (
	"bytes"
	"fmt"
	"github.com/cheekybits/is"
	"io/ioutil"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	s := NewShell(shellUrl)
	files := []string{"about", "readme"}
	for _, f := range files {
		file, err := os.Open(fmt.Sprintf("./testdata/%s", f))
		if err != nil {
			os.Exit(1)
		}

		err = s.FilesWrite(fmt.Sprintf("/testdata/%s", f), file, FilesWrite.Parents(true), FilesWrite.Create(true))
		if err != nil {
			os.Exit(1)
		}
	}

	exitVal := m.Run()
	if err := s.FilesRm("/testdata", true); err != nil {
		os.Exit(1)
	}
	os.Exit(exitVal)
}

func TestFilesChcid(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	err := s.FilesChcid("/testdata", FilesChcid.Hash("sha3-256"))
	is.Nil(err)

	stat, err := s.FilesStat("/testdata")
	is.Nil(err)
	is.Equal(stat.Hash, "bafybmigo44bvq5f4u2oswr7cilvlilftjekr4iilwxuxjj326hchztmk2m")

	err = s.FilesChcid("/testdata", FilesChcid.CidVersion(0))
	is.Nil(err)

	stat, err = s.FilesStat("/testdata")
	is.Nil(err)
	is.Equal(stat.Hash, "QmfZtacPc5nch976ZsiBw6nhLmTzy5JjW2pzZg8j7GjqWq")
}

func TestFilesCp(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	err := s.FilesCp("/testdata/readme", "/testdata/readme2")
	is.Nil(err)

	stat, err := s.FilesStat("/testdata/readme2")
	is.Nil(err)
	is.Equal(stat.Hash, "QmfZt7xPekp7npSM6DHDUnFseAiNZQs7wq6muH9o99RsCB")

	err = s.FilesRm("/testdata/readme2", true)
	is.Nil(err)
}

func TestFilesFlush(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	cid, err := s.FilesFlush("/testdata")
	is.Nil(err)
	is.Equal(cid, "QmfZtacPc5nch976ZsiBw6nhLmTzy5JjW2pzZg8j7GjqWq")
}

func TestFilesLs(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	list, err := s.FilesLs("/testdata", FilesLs.Long(true), FilesLs.U(true))
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

	err := s.FilesMkdir("/testdata/dir1/dir2", FilesMkdir.Parents(true), FilesMkdir.CidVersion(1), FilesMkdir.Hash("sha3-256"))
	is.Nil(err)

	err = s.FilesMkdir("/testdata/dir3/dir4")
	is.NotNil(err)

	err = s.FilesRm("/testdata/dir1", true)
	is.Nil(err)
}

func TestFilesMv(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	err := s.FilesMv("/testdata/readme", "/testdata/readme2")
	is.Nil(err)

	stat, err := s.FilesStat("/testdata/readme2")
	is.Nil(err)
	is.Equal(stat.Hash, "QmfZt7xPekp7npSM6DHDUnFseAiNZQs7wq6muH9o99RsCB")

	stat, err = s.FilesStat("/testdata/readme")
	is.NotNil(err)

	err = s.FilesMv("/testdata/readme2", "/testdata/readme")
	is.Nil(err)
}

func TestFilesRead(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	reader, err := s.FilesRead("/testdata/readme", FilesRead.Offset(0), FilesRead.Count(5))
	is.Nil(err)

	resBytes, err := ioutil.ReadAll(reader)
	is.Nil(err)
	is.Equal(string(resBytes), "Hello")
}

func TestFilesRm(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	file, _ := ioutil.ReadFile("./testdata/ping")
	err := s.FilesWrite("/testdata/dir1/ping", bytes.NewBuffer(file), FilesWrite.Parents(true), FilesWrite.Create(true))
	is.Nil(err)

	err = s.FilesRm("/testdata/dir1", false)
	is.NotNil(err)

	err = s.FilesRm("/testdata/dir1", true)
	is.Nil(err)
}

func TestFilesStat(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	res, err := s.FilesStat("/testdata")
	is.Nil(err)
	is.Equal(res.Hash, "QmfZtacPc5nch976ZsiBw6nhLmTzy5JjW2pzZg8j7GjqWq")
	is.Equal(res.Size, 0)
	is.Equal(res.Type, "directory")

	res, err = s.FilesStat("/testdata", FilesStat.WithLocal(true))
	is.Nil(err)
	is.Equal(res.WithLocality, true)
	is.Equal(res.Local, true)
	is.Equal(res.SizeLocal, 2997)
}

func TestFilesWrite(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	file, err := ioutil.ReadFile("./testdata/ping")
	is.Nil(err)

	err = s.FilesWrite("/testdata/ping", bytes.NewBuffer(file), FilesWrite.Create(true), FilesWrite.RawLeaves(true), FilesWrite.CidVersion(1), FilesWrite.Hash("sha3-256"))
	is.Nil(err)

	reader, err := s.FilesRead("/testdata/ping")
	is.Nil(err)

	resBytes, err := ioutil.ReadAll(reader)
	is.Nil(err)
	is.Equal(string(resBytes), "ipfs")

	file, err = ioutil.ReadFile("./testdata/ping")
	err = s.FilesWrite("/testdata/ping", bytes.NewBuffer(file), FilesWrite.Offset(0), FilesWrite.Count(2), FilesWrite.Truncate(true))
	is.Nil(err)

	reader, err = s.FilesRead("/testdata/ping")
	is.Nil(err)

	resBytes, err = ioutil.ReadAll(reader)
	is.Nil(err)
	is.Equal(string(resBytes), "ip")
}
