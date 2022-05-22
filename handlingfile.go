package zipgoserve

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (zf *ZipFileServer) GetHandlingFunction(mime string, file *zip.File, fm time.Time) func(w http.ResponseWriter, r *http.Request) {
	lm := fm.Format(http.TimeFormat)
	if int64(file.CompressedSize64) < zf.CachingThreshold {
		return zf.GetHandlingMemFunction(mime, file, fm)
	}
	if file.Method == deflateMethod {
		return func(w http.ResponseWriter, r *http.Request) {
			if ifModSince, err := http.ParseTime(r.Header.Get("if-modified-since")); err == nil {
				if !fm.IsZero() && fm.Before(ifModSince) {
					//log.Printf("Relpy unchanged > %s", file.Name)
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
			if zf.Locker != nil {
				zf.Locker.Lock()
				defer zf.Locker.Unlock()
			}
			//if r.Header.Get()
			if isDeflateAllowed(r) {
				rdf, err := file.OpenRaw()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				//log.Printf("Relpy compressed > %s", file.Name)
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
				log.Printf("Relpy uncompressed > %s", file.Name)
				w.Header().Set("Last-Modified", lm)
				r.Header.Add("Content-Length", strconv.Itoa(int(file.UncompressedSize64)))
				w.Header().Set("Content-Type", mime)
				io.Copy(w, rdf)
				rdf.Close()
			}

		}

	} else {
		return func(w http.ResponseWriter, r *http.Request) {
			if ifModSince, err := http.ParseTime(r.Header.Get("if-modified-since")); err == nil {
				if !fm.IsZero() && fm.Before(ifModSince) {
					//log.Printf("Relpy unchanged  %s", file.Name)
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
			{
				if zf.Locker != nil {
					zf.Locker.Lock()
					defer zf.Locker.Unlock()
				}

				rdf, err := file.Open()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				//log.Printf("Relpy uncompressed  %s", file.Name)
				w.Header().Set("Last-Modified", lm)
				r.Header.Add("Content-Length", strconv.Itoa(int(file.UncompressedSize64)))
				w.Header().Set("Content-Type", mime)
				io.Copy(w, rdf)
				rdf.Close()
			}

		}

	}

}
