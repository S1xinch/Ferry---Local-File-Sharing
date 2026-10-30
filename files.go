package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const maxUpload = 8 << 30 // 8 GB per request

// fileShare serves one folder. os.Root makes it impossible for a request
// to reach anything outside that folder, even with tricks like "../".
type fileShare struct {
	root     *os.Root
	name     string
	readOnly bool
}

func newFileShare(dir string, readOnly bool) (*fileShare, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot open folder: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is a file, not a folder", dir)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	return &fileShare{root: root, name: filepath.Base(dir), readOnly: readOnly}, nil
}

// rel turns a URL path into a path inside the shared folder ("." for the top).
func rel(urlPath string) string {
	p := strings.TrimPrefix(path.Clean("/"+urlPath), "/")
	if p == "" {
		return "."
	}
	return p
}

func (s *fileShare) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := rel(r.URL.Path)
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		s.serveGet(w, r, p)
	case http.MethodPost:
		if s.readOnly {
			http.Error(w, "Uploads are turned off for this share.", http.StatusForbidden)
			return
		}
		s.upload(w, r, p)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *fileShare) serveGet(w http.ResponseWriter, r *http.Request, p string) {
	f, err := s.root.Open(p)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if !info.IsDir() {
		if r.URL.Query().Has("dl") {
			w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(info.Name()))
		}
		http.ServeContent(w, r, info.Name(), info.ModTime(), f)
		return
	}
	if !strings.HasSuffix(r.URL.Path, "/") {
		http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
		return
	}
	if r.URL.Query().Has("zip") {
		s.zipFolder(w, p)
		return
	}

	entries, err := f.ReadDir(-1)
	if err != nil {
		http.Error(w, "Cannot read folder", http.StatusInternalServerError)
		return
	}
	s.list(w, r, p, entries)
}

// Image files get a real thumbnail. Bigger ones get a type tile instead, so a
// folder of camera originals still opens quickly.
const maxThumb = 30 << 20

var thumbExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}

type entry struct {
	Name, Href, Size, Modified string
	Badge                      string // short type label for the tile, e.g. "PDF"
	IsDir, Thumb               bool
}

// badge is the label on a file's tile: the extension in capitals, or FILE.
func badge(ext string) string {
	ext = strings.TrimPrefix(ext, ".")
	if ext == "" || len(ext) > 4 {
		return "FILE"
	}
	return strings.ToUpper(ext)
}

type crumb struct{ Name, Href string }

type listPage struct {
	Title    string
	Crumbs   []crumb
	Entries  []entry
	ReadOnly bool
	Notice   string
}

func (s *fileShare) list(w http.ResponseWriter, r *http.Request, p string, des []os.DirEntry) {
	page := listPage{Title: s.name, ReadOnly: s.readOnly}
	switch r.URL.Query().Get("done") {
	case "":
	case "1":
		page.Notice = "1 file uploaded."
	default:
		page.Notice = r.URL.Query().Get("done") + " files uploaded."
	}

	page.Crumbs = []crumb{{Name: s.name, Href: "/"}}
	if p != "." {
		href := "/"
		for _, part := range strings.Split(p, "/") {
			href += url.PathEscape(part) + "/"
			page.Crumbs = append(page.Crumbs, crumb{Name: part, Href: href})
		}
	}

	sort.Slice(des, func(i, j int) bool {
		if des[i].IsDir() != des[j].IsDir() {
			return des[i].IsDir()
		}
		return strings.ToLower(des[i].Name()) < strings.ToLower(des[j].Name())
	})
	for _, d := range des {
		info, err := d.Info()
		if err != nil {
			continue
		}
		e := entry{Name: d.Name(), Href: url.PathEscape(d.Name()), IsDir: d.IsDir(), Modified: info.ModTime().Format("2 Jan 2006 15:04")}
		if d.IsDir() {
			e.Href += "/"
			e.Size = "folder"
		} else {
			e.Size = humanSize(info.Size())
			ext := strings.ToLower(path.Ext(d.Name()))
			e.Thumb = thumbExts[ext] && info.Size() <= maxThumb
			e.Badge = badge(ext)
		}
		page.Entries = append(page.Entries, e)
	}
	render(w, "list", page)
}

func (s *fileShare) upload(w http.ResponseWriter, r *http.Request, dir string) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "Bad upload", http.StatusBadRequest)
		return
	}
	count := 0
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "Upload interrupted: "+err.Error(), http.StatusBadRequest)
			return
		}
		name := filepath.Base(filepath.Clean("/" + strings.ReplaceAll(part.FileName(), `\`, "/")))
		if part.FormName() != "files" || name == "" || name == "/" || name == "." {
			part.Close()
			continue
		}
		target := s.freeName(dir, name)
		out, err := s.root.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			part.Close()
			http.Error(w, "Cannot save "+name, http.StatusInternalServerError)
			return
		}
		_, copyErr := io.Copy(out, part)
		out.Close()
		part.Close()
		if copyErr != nil {
			s.root.Remove(target)
			http.Error(w, "Upload interrupted", http.StatusBadRequest)
			return
		}
		logf("Received %s from %s", target, clientIP(r))
		count++
	}
	http.Redirect(w, r, r.URL.Path+fmt.Sprintf("?done=%d", count), http.StatusSeeOther)
}

// freeName avoids overwriting: "photo.jpg" becomes "photo (1).jpg" if taken.
func (s *fileShare) freeName(dir, name string) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	candidate := name
	for i := 1; ; i++ {
		full := path.Join(dir, candidate)
		if _, err := s.root.Stat(full); os.IsNotExist(err) {
			return full
		}
		candidate = fmt.Sprintf("%s (%d)%s", base, i, ext)
	}
}

func (s *fileShare) zipFolder(w http.ResponseWriter, p string) {
	name := s.name
	if p != "." {
		name = path.Base(p)
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(name+".zip"))
	zw := zip.NewWriter(w)
	defer zw.Close()
	fsys := s.root.FS()
	fs.WalkDir(fsys, p, func(fp string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !d.Type().IsRegular() {
			return nil
		}
		relName := fp
		if p != "." {
			relName = strings.TrimPrefix(fp, p+"/")
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		hdr, _ := zip.FileInfoHeader(info)
		hdr.Name = relName
		hdr.Method = zip.Deflate
		dst, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		src, err := fsys.Open(fp)
		if err != nil {
			return nil
		}
		defer src.Close()
		_, err = io.Copy(dst, src)
		return err
	})
}

func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGT"[exp])
}
