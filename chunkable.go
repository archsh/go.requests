package requests

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"os"
	"strconv"
)

const defaultChunkSize = 1024 * 1024 * 2
const defaultConcurrents = 5
const defaultRetries = 3

type FileChunk struct {
	RefFile      string
	TotalChunks  uint
	TotalMD5     string
	TotalSize    uint64
	ChunkNum     uint
	ChunkSize    uint64
	ChunkMD5     string
	_bodyBuffer  *bytes.Buffer
	_contentType string
}

func (fc FileChunk) Query() url.Values {
	v := url.Values{}
	v.Add("refFile", fc.RefFile)
	v.Add("totalChunks", fmt.Sprint(fc.TotalChunks))
	v.Add("totalMD5", fc.TotalMD5)
	v.Add("totalSize", fmt.Sprint(fc.TotalSize))
	v.Add("chunkNum", fmt.Sprint(fc.ChunkNum))
	v.Add("chunkSize", fmt.Sprint(fc.ChunkSize))
	v.Add("chunkMD5", fc.ChunkMD5)
	return v
}

func (fc FileChunk) URL(s string) string {
	u, _ := url.Parse(s)
	for k, v := range fc.Query() {
		for _, vv := range v {
			u.Query().Add(k, vv)
		}
	}
	return u.String()
}

func (fc FileChunk) Upload(s string) error {
	var us = fc.URL(s)
	if resp, e := Get(us); nil == e && resp._resp.StatusCode/100 == 2 {
		return nil
	}
	if resp, e := builtinClient.Post(us, fc._contentType, fc._bodyBuffer); nil != e {
		return nil
	} else if resp.StatusCode/100 != 2 {
		return fmt.Errorf("%s", resp.Status)
	}
	return nil
}

func ParseChunk(vals url.Values) FileChunk {
	var fc = FileChunk{TotalChunks: 1}
	fc.RefFile = vals.Get("refFile")
	fc.ChunkMD5 = vals.Get("chunkMD5")
	fc.TotalMD5 = vals.Get("totalMD5")

	//TotalChunks  uint
	if n, e := strconv.ParseUint(vals.Get("totalChunks"), 10, 64); nil == e {
		fc.TotalChunks = uint(n)
	}
	//TotalSize    uint64
	if n, e := strconv.ParseUint(vals.Get("totalSize"), 10, 64); nil == e {
		fc.TotalSize = n
	}
	//ChunkNum     uint
	if n, e := strconv.ParseUint(vals.Get("chunkNum"), 10, 64); nil == e {
		fc.ChunkNum = uint(n)
	}
	//ChunkSize    uint64
	if n, e := strconv.ParseUint(vals.Get("chunkSize"), 10, 64); nil == e {
		fc.ChunkSize = n
	}

	return fc
}

func UploadFile(s string, fp *os.File) error {
	var totalSize uint64
	var refFile string
	var totalChunks int
	if st, e := fp.Stat(); nil != e {
		return e
	} else if st.IsDir() {
		return fmt.Errorf("can not upload a directory")
	} else {
		totalSize = uint64(st.Size())
		refFile = st.Name()
	}
	if (totalSize % defaultChunkSize) == 0 {
		totalChunks = int(totalSize / defaultChunkSize)
	} else {
		totalChunks = int(totalSize/defaultChunkSize + 1)
	}
	for i := 0; i < totalChunks; i++ {
		var chunkMd5 string
		bodyBuffer := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuffer)
		bs := make([]byte, defaultChunkSize)
		n, e := fp.Read(bs)
		if nil != e {
			if e != io.EOF {
				return e
			}
		}
		chunkMd5 = hex.EncodeToString(md5.New().Sum(bs))
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				"file", refFile))
		h.Set("Content-Type", "application/octet-stream")
		if fileWriter, e := bodyWriter.CreatePart(h); nil != e {
			return e
		} else {
			if nn, ee := fileWriter.Write(bs); nil != ee {
				return ee
			} else if nn != n {
				return fmt.Errorf("write bytes not matched: %d <> %d", n, nn)
			}
		}
		fc := FileChunk{
			RefFile:      refFile,
			ChunkNum:     uint(i),
			ChunkSize:    uint64(n),
			ChunkMD5:     chunkMd5,
			TotalSize:    totalSize,
			TotalChunks:  uint(totalChunks),
			TotalMD5:     "",
			_bodyBuffer:  bodyBuffer,
			_contentType: bodyWriter.FormDataContentType(),
		}
		if e := bodyWriter.Close(); nil != e {
			return e
		}
		if e := fc.Upload(s); nil != e {
			return e
		}
	}
	return nil
}
