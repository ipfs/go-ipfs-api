package shell

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	files "github.com/whyrusleeping/go-multipart-files"
)

func (s *Shell) DagPut(data, ienc, kind string) (string, error) {
	req := s.newRequest("dag/put")
	req.Opts = map[string]string{
		"input-enc": ienc,
		"format":    kind,
	}

	r := strings.NewReader(data)
	rc := ioutil.NopCloser(r)
	fr := files.NewReaderFile("", "", rc, nil)
	slf := files.NewSliceFile("", "", []files.File{fr})
	fileReader := files.NewMultiFileReader(slf, true)
	req.Body = fileReader

	resp, err := req.Send(s.httpcli)
	if err != nil {
		return "", err
	}
	defer resp.Close()

	if resp.Error != nil {
		return "", resp.Error
	}

	var out struct {
		Cid struct {
			Target string `json:"/"`
		}
	}
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return "", err
	}

	return out.Cid.Target, nil
}
