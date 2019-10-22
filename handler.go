package requests

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
)

var (
	tmpMapMtx sync.Mutex
	tmpMap    map[string]uint
)

func tmpCounter(k string) uint {
	tmpMapMtx.Lock()
	defer tmpMapMtx.Unlock()
	n, _ := tmpMap[k]
	n += 1
	return n
}
func tmpForget(k string) {
	tmpMapMtx.Lock()
	defer tmpMapMtx.Unlock()
	delete(tmpMap, k)
}

func makeTempDir(temp, filename string) string {
	return path.Join(temp, hex.EncodeToString(md5.New().Sum([]byte(filename))))
}

func makeTempFile(temp, filename string, chunkNum uint) string {
	return path.Join(makeTempDir(temp, filename), fmt.Sprintf("%d.dat", chunkNum))
}

func fileServe(root, temp, filename string, fc FileChunk, w http.ResponseWriter, r *http.Request) error {
	if _, e := os.Stat(path.Join(root, filename)); nil == e {
		http.ServeFile(w, r, path.Join(root, filename))
	} else if fc.TotalChunks > 1 {
		tempFile := makeTempFile(temp, filename, fc.ChunkNum)
		if st, e := os.Stat(tempFile); nil == e && uint64(st.Size()) == fc.ChunkSize {
			http.ServeFile(w, r, tempFile)
		}
	}
	http.NotFound(w, r)
	return nil
}

func uploadFile(root, temp, filename string, fc FileChunk, w http.ResponseWriter, r *http.Request, done func()) error {
	if file, header, err := r.FormFile("file"); err != nil {
		//log.Errorln("UploadFile:> get upload file failed:", err)
		return err
	} else {
		defer file.Close()
		if filename == "" {
			filename = header.Filename
		}

		bs, e := ioutil.ReadAll(file)
		if nil != e {
			return e
		}
		if fc.ChunkMD5 != "" {
			if fc.ChunkMD5 != hex.EncodeToString(md5.New().Sum(bs)) {
				return fmt.Errorf("checksum failed")
			}
		}

		if fc.TotalChunks > 1 {
			k := makeTempDir(temp, filename)
			tmpFile := makeTempFile(temp, filename, fc.ChunkNum)
			var chunks uint
			if fp, e := os.Create(tmpFile); nil != e {
				return e
			} else {
				defer fp.Close()
				if n, e := fp.Write(bs); nil != e {
					return e
				} else if n != len(bs) {
					return fmt.Errorf("writen bytes not match uploaded")
				} else {
					chunks = tmpCounter(k)
					if chunks >= fc.TotalChunks { // Merge segments
						dest := path.Join(root, filename)
						if e := os.MkdirAll(path.Dir(dest), 0755); nil != e {
							return e
						}
						if fp, e := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, 0644); nil != e {
							//log.Errorln("UploadFile:> Create file failed:", e)
							return e
						} else {
							defer fp.Close()
							for i := uint(0); i < fc.TotalChunks; i++ {
								segfile := makeTempFile(temp, filename, i)
								if sf, e := os.Open(segfile); nil != e {
									return e
								} else if _, e := io.Copy(fp, sf); nil != e {
									sf.Close()
									return e
								} else {
									sf.Close()
								}
							}
							_ = os.RemoveAll(makeTempDir(temp, filename))
							tmpForget(k)
							done()
						}
					}
				}
			}
		} else {
			fullname := path.Join(root, filename)
			if fp, e := os.Create(fullname); nil != e {
				return e
			} else {
				defer fp.Close()
				if n, e := fp.Write(bs); nil != e {
					return e
				} else if n != len(bs) {
					return fmt.Errorf("writen bytes not match uploaded")
				} else {
					done()
				}
			}
		}
	}
	return nil
}

func UploadHandlerFunc(root, temp, filename string, w http.ResponseWriter, r *http.Request, done func()) error {
	fc := ParseChunk(r.URL.Query())
	if filename == "" {
		filename = fc.RefFile
	}
	if strings.ToUpper(r.Method) == "GET" || strings.ToUpper(r.Method) == "HEAD" {
		return fileServe(root, temp, filename, fc, w, r)
	} else if strings.ToUpper(r.Method) == "POST" {
		return uploadFile(root, temp, filename, fc, w, r, done)
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return fmt.Errorf("not allowed")
	}
	return nil
}

func init() {
	tmpMap = make(map[string]uint)
}
