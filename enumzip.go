package zipgoserve

import (
	"archive/zip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

type ZipFileServer struct {
	r                *zip.ReadCloser
	HTTPprefix       string
	IndexSuffix      string
	PATHprefix       string // html/
	CachingThreshold int64
	Mime             []MIMEMap
	Locker           *sync.Mutex // Looks like zip is concurrent read safe - locking removed
}

type MIMEMap struct {
	Suffix string
	MIME   string
}

func HardcodedMap() []MIMEMap {
	return []MIMEMap{
		{".html", "text/html"},
		{".htm", "text/html"},
		{".xml", "text/xml"},
		{".txt", "text/plain"},
		{".jpg", "image/jpeg"},
		{".jpeg", "image/jpeg"},
		{".png", "image/png"},
		{".gif", "image/gif"},
		{".css", "text/css"},
		{".js", "text/javascript"},
		{".map", "application/json"},
		{".json", "application/json"},
	}
}

const mimePATH = "mime.json"
const indexName = "index.html"

func (zf *ZipFileServer) AttachFile(fn string) {
	//zf.name = fn
	var err error
	zf.r, err = zip.OpenReader(fn)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

func (zf *ZipFileServer) Close() error {
	return zf.r.Close()
}

func (zf *ZipFileServer) ParseJSONMIME() error {
	for _, f := range zf.r.File {
		if f.Name != mimePATH {
			continue
		}

		// Found it, print its content to terminal:
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}

		buf, err := io.ReadAll(rc)

		err = json.Unmarshal(buf, &zf.Mime)
		if err != nil {
			return err
		}
		break
	}
	return nil
}

func isDeflateAllowed(r *http.Request) bool {
	ae := r.Header.Get("Accept-Encoding")
	sae := strings.Split(ae, ", ")
	for _, v := range sae {
		if strings.HasPrefix(v, "deflate") {
			return true
		}
	}
	return false
}

const deflateMethod = 8

func (zf *ZipFileServer) MapFiles(mux *http.ServeMux) error {
	for _, f := range zf.r.File {
		mmOK := false
		if strings.HasPrefix(f.Name, zf.PATHprefix) && !f.FileInfo().IsDir() {
			for _, mm := range zf.Mime {
				if strings.HasSuffix(f.Name, mm.Suffix) {
					URL := zf.HTTPprefix + f.Name[len(zf.PATHprefix):]
					log.Printf("Mapping file %s to URL %s as %s (compression %d)", f.Name, URL, mm.MIME, f.Method)
					fMap := zf.GetHandlingFunction(mm.MIME, f, f.Modified)
					mux.HandleFunc(URL, fMap)
					if strings.HasSuffix(URL, zf.IndexSuffix) {
						URLi := URL[:len(URL)-len(zf.IndexSuffix)]
						mux.HandleFunc(URLi, fMap)
					}
					mmOK = true
					break
				}
			}

		}
		if !mmOK {
			log.Printf("Ignoring file %s", f.Name)
		}

	}
	return nil
}
