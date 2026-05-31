package main

import (
	"archive/zip"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	webRoot = "/var/www/site"
	maxSize = 100 << 20 // 100MB
)

func main() {
	authToken := os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		log.Fatal("AUTH_TOKEN is required")
	}

	http.HandleFunc("/_theres_no_way_you_have_this_in_your_static_site", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if !authorized(r, authToken) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxSize)
		if err := r.ParseMultipartForm(maxSize); err != nil {
			http.Error(w, "invalid upload", http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "file missing", http.StatusBadRequest)
			return
		}
		defer file.Close()

		tmpZip, err := os.CreateTemp("", "site-*.zip")
		if err != nil {
			http.Error(w, "server error", 500)
			return
		}
		defer os.Remove(tmpZip.Name())

		if _, err := io.Copy(tmpZip, file); err != nil {
			http.Error(w, "upload failed", 500)
			return
		}

		if err := extractZip(tmpZip.Name(), webRoot); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("site deployed\n"))
	})

	log.Println("Upload server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func authorized(r *http.Request, token string) bool {
	auth := r.Header.Get("Authorization")
	return auth == "Bearer "+token
}

func extractZip(zipPath, dest string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	os.RemoveAll(dest)
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)

		cleanDest := filepath.Clean(dest) + string(os.PathSeparator)
		cleanTarget := filepath.Clean(target)

		if !strings.HasPrefix(cleanTarget, cleanDest) {
			return errors.New("invalid zip path")
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(cleanTarget, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(cleanTarget), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		out, err := os.Create(cleanTarget)
		if err != nil {
			rc.Close()
			return err
		}

		_, copyErr := io.Copy(out, rc)

		rc.Close()
		out.Close()

		if copyErr != nil {
			return copyErr
		}
	}

	return nil
}

