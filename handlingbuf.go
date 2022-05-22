package zipgoserve

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (zf *ZipFileServer) GetHandlingMemFunction(mime string, file *zip.File, fm time.Time) func(w http.ResponseWriter, r *http.Request) {
	lm := fm.Format(http.TimeFormat)
	if file.Method == deflateMethod {
		var buf []byte
		{
			rdf, err := file.OpenRaw()
			if err != nil {
				log.Fatal(err)
				return nil
			}
			buf, err = io.ReadAll(rdf)
			if err != nil {
				log.Fatal(err)
				return nil
			}
		}
		return func(w http.ResponseWriter, r *http.Request) {
			if ifModSince, err := http.ParseTime(r.Header.Get("if-modified-since")); err == nil {
				if !fm.IsZero() && fm.Before(ifModSince) {
					//log.Printf("Relpy unchanged >> %s", file.Name)
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
			//if r.Header.Get()
			if isDeflateAllowed(r) {
				//log.Printf("Relpy compressed >> %s", file.Name)
				w.Header().Set("Last-Modified", lm)
				r.Header.Add("Content-Length", strconv.Itoa(int(file.CompressedSize64)))
				w.Header().Set("Content-Type", mime)
				w.Header().Set("Content-Encoding", "deflate")
				w.Write(buf)
			} else {
				fr := flate.NewReader(bytes.NewReader(buf))
				log.Printf("Relpy uncompressed > %s", file.Name)
				w.Header().Set("Last-Modified", lm)
				r.Header.Add("Content-Length", strconv.Itoa(int(file.UncompressedSize64)))
				w.Header().Set("Content-Type", mime)
				io.Copy(w, fr)
				fr.Close()
			}
		}

	} else {
		var buf []byte
		{
			rdf, err := file.Open()
			if err != nil {
				log.Fatal(err)
				return nil
			}
			buf, err = io.ReadAll(rdf)
			if err != nil {
				log.Fatal(err)
				return nil
			}
		}

		return func(w http.ResponseWriter, r *http.Request) {
			if ifModSince, err := http.ParseTime(r.Header.Get("if-modified-since")); err == nil {
				if !fm.IsZero() && fm.Before(ifModSince) {
					log.Printf("Relpy unchanged  # %s", file.Name)
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
			{
				log.Printf("Relpy uncompressed  # %s", file.Name)
				w.Header().Set("Last-Modified", lm)
				r.Header.Add("Content-Length", strconv.Itoa(int(file.UncompressedSize64)))
				w.Header().Set("Content-Type", mime)
				w.Write(buf)
			}

		}

	}

}
