package main

import (
	"archive/zip"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Gzip middleware
func gzipHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzw, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func main() {
	authToken := os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		log.Fatal("AUTH_TOKEN environment variable not set")
	}

	staticDir := "./site"
	os.MkdirAll(staticDir, os.ModePerm)

	uploadPath := "/_theres_no_way_you_have_this_in_your_static_site"

	// Upload endpoint (requires token)
	http.HandleFunc(uploadPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}
		token := r.Header.Get("Authorization")
		if token != "Bearer "+authToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		file, _, err := r.FormFile("site")
		if err != nil {
			http.Error(w, "Failed to read file: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Clear old site
		os.RemoveAll(staticDir)
		os.MkdirAll(staticDir, os.ModePerm)

		tmpZip := "site.zip"
		out, err := os.Create(tmpZip)
		if err != nil {
			http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer out.Close()
		io.Copy(out, file)

		err = unzip(tmpZip, staticDir)
		if err != nil {
			http.Error(w, "Failed to unzip: "+err.Error(), http.StatusInternalServerError)
			return
		}
		os.Remove(tmpZip)

		w.Write([]byte("Site uploaded successfully!"))
	})

	// Serve static files with gzip
	fs := http.FileServer(http.Dir(staticDir))
	http.Handle("/", gzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(staticDir, r.URL.Path)
		info, err := os.Stat(path)
		if os.IsNotExist(err) || info.IsDir() {
			// fallback to index.html
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	})))

	log.Printf("Server running on port 80\nUpload path: %s\n", uploadPath)
	log.Fatal(http.ListenAndServe(":80", nil))
}

// unzip helper
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Remove top-level directory if present
		fname := f.Name
		if parts := strings.SplitN(f.Name, "/", 2); len(parts) == 2 {
			fname = parts[1]
		}

		fpath := filepath.Join(dest, fname)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		os.MkdirAll(filepath.Dir(fpath), os.ModePerm)

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
