package zipgoserve

import (
	"archive/zip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ZipFileServer struct {
	r           *zip.ReadCloser
	HTTPprefix  string
	IndexSuffix string
	PATHprefix  string // html/
	Mime        []MIMEMap
	ziplock     *sync.Mutex // Looks like zip is concurrent read safe - locking removed
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

func (zf *ZipFileServer) MapFile(mux *http.ServeMux, mime string, url string, file *zip.File, fm time.Time) bool {
	lm := fm.Format(http.TimeFormat)
	log.Printf("Mapping file %s to URL %s as %s (compression %d)", file.Name, url, mime, file.Method)
	if file.Method != 0 {
		mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
			if ifModSince, err := http.ParseTime(r.Header.Get("if-modified-since")); err == nil {
				if !fm.IsZero() && fm.Before(ifModSince) {
					log.Printf("Relpy unchanged > %s", url)
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
			if zf.ziplock != nil {
				zf.ziplock.Lock()
				defer zf.ziplock.Unlock()
			}
			//if r.Header.Get()
			if isDeflateAllowed(r) {
				rdf, err := file.OpenRaw()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				log.Printf("Relpy compressed > %s", url)
				w.Header().Set("Last-Modified", lm)
				r.Header.Add("Content-Length", strconv.Itoa(int(file.CompressedSize64)))
				w.Header().Set("Content-Type", mime)
				w.Header().Set("Content-Encoding", "deflate")
				io.Copy(w, rdf)
			} else {
				rdf, err := file.Open()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				log.Printf("Relpy uncompressed > %s", url)
				w.Header().Set("Last-Modified", lm)
				r.Header.Add("Content-Length", strconv.Itoa(int(file.UncompressedSize64)))
				w.Header().Set("Content-Type", mime)
				io.Copy(w, rdf)
				rdf.Close()
			}

		})

	} else {
		mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
			if ifModSince, err := http.ParseTime(r.Header.Get("if-modified-since")); err == nil {
				if !fm.IsZero() && fm.Before(ifModSince) {
					log.Printf("Relpy unchanged  %s", url)
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
			{
				if zf.ziplock != nil {
					zf.ziplock.Lock()
					defer zf.ziplock.Unlock()
				}

				rdf, err := file.Open()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				log.Printf("Relpy uncompressed  %s", url)
				w.Header().Set("Last-Modified", lm)
				r.Header.Add("Content-Length", strconv.Itoa(int(file.UncompressedSize64)))
				w.Header().Set("Content-Type", mime)
				io.Copy(w, rdf)
				rdf.Close()
			}

		})

	}
	return true
}

func (zf *ZipFileServer) MapFiles(mux *http.ServeMux) error {
	for _, f := range zf.r.File {
		mmOK := false
		if strings.HasPrefix(f.Name, zf.PATHprefix) && !f.FileInfo().IsDir() {
			for _, mm := range zf.Mime {
				if strings.HasSuffix(f.Name, mm.Suffix) {
					URL := zf.HTTPprefix + f.Name[len(zf.PATHprefix):]
					if strings.HasSuffix(URL, zf.IndexSuffix) {
						URLi := URL[:len(URL)-len(zf.IndexSuffix)]
						mmOK = zf.MapFile(mux, mm.MIME, URLi, f, f.Modified)
					}
					mmOK = (zf.MapFile(mux, mm.MIME, URL, f, f.Modified) || mmOK)
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
