package main

import (
	"github.com/msinev/zipgoserve"
	"net/http"
	"os"
	"sync"
)

func main() {
	zf := &zipgoserve.ZipFileServer{
		HTTPprefix:       "/",
		IndexSuffix:      "index.html",
		PATHprefix:       "html/",
		Mime:             zipgoserve.HardcodedMap(),
		CachingThreshold: 1024,
		Locker:           &sync.Mutex{},
	}

	if len(os.Args) != 3 {
		println("exec [ZIP File] [bind]")
		os.Exit(-1)
		return
	}

	muxer := &http.ServeMux{}
	zf.AttachFile(os.Args[1])
	zf.MapFiles(muxer)
	http.ListenAndServe(os.Args[2], muxer)

}
