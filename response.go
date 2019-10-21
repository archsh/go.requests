package requests

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"

	yaml "gopkg.in/yaml.v2"
)

type Response struct {
	_resp *http.Response
}

func (resp Response) Bytes() ([]byte, error) {
	return ioutil.ReadAll(resp._resp.Body)
}

func (resp Response) JSON(o interface{}) error {
	if bs, e := resp.Bytes(); nil != e {
		return e
	} else if e := json.Unmarshal(bs, o); nil != e {
		return e
	}
	return nil
}

func (resp Response) YAML(o interface{}) error {
	if bs, e := resp.Bytes(); nil != e {
		return e
	} else if e := yaml.Unmarshal(bs, o); nil != e {
		return e
	}
	return nil
}

func (resp Response) XML(o interface{}) error {
	if bs, e := resp.Bytes(); nil != e {
		return e
	} else if e := xml.Unmarshal(bs, o); nil != e {
		return e
	}
	return nil
}
