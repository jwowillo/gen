// Package server has a constructor for a http.Handler which serves static files
// from a directory and properly sets or doesn't set the Gzip header for each
// file.
package server

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
)

// Handler that serves static files from the passed directory and properly sets
// or doesn't set the Gzip header.
func Handler(dir string) http.Handler {
	h := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(dir, r.URL.String())
		if filepath.Ext(path) == "" {
			path = filepath.Join(path, "index.html")
		}
		f, err := os.Open(path)
		if err != nil {
			h.ServeHTTP(w, r)
			return
		}
		defer f.Close()
		rd := bufio.NewReader(f)
		bs, err := rd.Peek(2)
		if err != nil {
			h.ServeHTTP(w, r)
			return
		}
		if bs[0] == 0x1f && bs[1] == 0x8b {
			w.Header().Set("Content-Encoding", "gzip")
		}
		h.ServeHTTP(w, r)
	})
}
