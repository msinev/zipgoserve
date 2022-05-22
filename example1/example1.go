package main

import (
	"github.com/msinev/zipgoserve"
	"net/http"
	"os"
)

func main() {
	zf := &zipserver.ZipFileServer{
		HTTPprefix:  "/",
		IndexSuffix: "index.html",
		PATHprefix:  "html/",
		Mime:        zipserver.HardcodedMap(),
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
