package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"
)

var (
	rootDir = os.Getenv("PT_DIR")

	authKey      = os.Getenv("PT_AUTH")
	serverAddr   = os.Getenv("PT_ADDR")
	specialCodes = []string{scRoot, scList, scFallback}
)

// Special codes
const (
	// .list is not a code file, but a special code to list all codes.
	scList = ".list"
	// root and fallback are actually code files, but can't be set via the API.
	// Set them manually on server. `.root` is the target for the root path.
	// `.fallback` is the target for non-existent codes.
	scRoot     = ".root"
	scFallback = ".fallback"
)

func err500(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 - server error\n"))
	return true
}

func auth(w http.ResponseWriter, r *http.Request) bool {
	if authKey == "" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("403 - listing disabled\n"))
		return false
	}
	if r.URL.Query().Get("auth") == authKey {
		return true
	}
	if r.Header.Get("auth") == authKey {
		return true
	}
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("401 - wrong auth key\n"))
	return false
}

type Link struct {
	Code   string `json:"code"`
	Target string `json:"target"`
}

func listPages(w http.ResponseWriter, r *http.Request) {
	if !auth(w, r) {
		return
	}
	d, err := os.ReadDir(rootDir)
	if err500(w, err) {
		return
	}
	var links []Link
	for _, f := range d {
		if f.IsDir() {
			continue
		}
		target, err := os.ReadFile(path.Join(rootDir, f.Name()))
		link := Link{Code: f.Name()}
		if err != nil {
			link.Target = "read error: " + err.Error()
		} else {
			link.Target = strings.TrimSpace(string(target))
		}
		links = append(links, link)
	}
	if r.URL.Query().Has("json") || r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		out, err := json.Marshal(links)
		if err != nil {
			err500(w, err)
		}
		w.Write(out)
	} else {
		for _, l := range links {
			w.Write([]byte(l.Code + "\n" + l.Target + "\n\n"))
		}
	}
}

func getFallback() string {
	target, err := os.ReadFile(path.Join(rootDir, scFallback))
	if err == nil {
		return strings.TrimSpace(string(target))
	}
	if !errors.Is(err, os.ErrNotExist) {
		log.Println("unexpected error when reading fallback:", err)
	}
	return ""
}

func get(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")
	if code == "" {
		code = scRoot
	}
	if code == scList {
		listPages(w, r)
		return
	}
	target, err := os.ReadFile(path.Join(rootDir, code))
	if err != nil {
		fallback := getFallback()
		if fallback != "" {
			http.Redirect(w, r, fallback, http.StatusFound)
			return
		}
		if errors.Is(err, os.ErrNotExist) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - not found\n"))
		} else {
			err500(w, err)
		}
		return
	}
	http.Redirect(w, r, strings.TrimSpace(string(target)), http.StatusFound)
}

func post(w http.ResponseWriter, r *http.Request) {
	if authKey == "" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !auth(w, r) {
		return
	}
	code := strings.TrimPrefix(r.URL.Path, "/")
	if slices.Contains(specialCodes, code) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("403 - cannot use special code\n"))
		return
	}
	if code == "" {
		var rc [3]byte
		_, err := rand.Read(rc[:])
		if err500(w, err) {
			return
		}
		code = base64.RawURLEncoding.EncodeToString(rc[:])
	}
	f := path.Join(rootDir, code)
	_, err := os.Stat(f)
	if !errors.Is(err, os.ErrNotExist) {
		if err500(w, err) {
			return
		}
		w.Write([]byte("403 - file already exists\n"))
		return
	}
	defer r.Body.Close()
	target, err := io.ReadAll(r.Body)
	if err500(w, err) {
		return
	}
	err = os.WriteFile(f, target, 0644)
	if err500(w, err) {
		return
	}
	w.Write([]byte(code + "\n"))
}

func handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		get(w, r)
	case http.MethodPost:
		post(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	addr := serverAddr
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	if addr == "" {
		addr = ":4600"
	}
	if rootDir == "" {
		log.Fatalln("PT_DIR is not set")
	}
	fi, err := os.Stat(rootDir)
	if err != nil {
		log.Fatalln("failed to stat root dir:", err)
	}
	if !fi.IsDir() {
		log.Fatalf("%v is not a directory", rootDir)
	}

	http.HandleFunc("/", handle)
	log.Printf("starting pt at %s", addr)
	http.ListenAndServe(addr, nil)
}
