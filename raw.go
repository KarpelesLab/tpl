package tpl

import (
	"archive/zip"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// RawData contains the raw (uncompiled) data for current template
type RawData struct {
	PageProperties map[string]string // typically Charset=UTF-8 Content_Type=text/html
	TemplateData   map[string]string // actual contents for templates, at least "main" should be there
}

// Init sets the raw storage, required before setting contents (if not loading from disk)
func (r *RawData) init() {
	r.PageProperties = make(map[string]string)
	r.TemplateData = make(map[string]string)

	// default values
	r.PageProperties["Charset"] = "UTF-8"
	r.PageProperties["Content-Type"] = "text/html"
}

func grabZipFile(f *zip.File) ([]byte, error) {
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// read all
	return io.ReadAll(r)
}

func grabVfsFile(fs fs.FS, path string) ([]byte, error) {
	r, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// read all
	return io.ReadAll(r)
}

func grabFile(path string) ([]byte, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// read all
	return io.ReadAll(r)
}

func (r *RawData) IsValid() bool {
	_, ok := r.TemplateData["main"]
	return ok
}

func (res *RawData) FromZip(z *zip.Reader, d ...string) error {
	for _, rd := range d {
		if !strings.HasSuffix(rd, "/") {
			rd += "/"
		}

		for _, f := range z.File {
			if strings.HasSuffix(f.Name, "/") {
				continue // ignore dir
			}
			if !strings.HasPrefix(f.Name, rd) {
				continue
			}
			localName := f.Name[len(rd):]

			data, err := grabZipFile(f)
			if err != nil {
				return err
			}

			if localName == "_properties.json" {
				var jsonDec map[string]string
				err = json.Unmarshal(data, &jsonDec)
				if err != nil {
					return err
				}

				// set data
				for k, v := range jsonDec {
					res.PageProperties[strings.Replace(k, "_", "-", -1)] = v
				}
				continue
			}

			if strings.HasSuffix(localName, ".tpl") {
				res.TemplateData[localName[:len(localName)-4]] = string(data)
				continue
			}
			log.Printf("[tpl] ignoring unknown file %s", f.Name)
		}
	}
	return nil
}

func (res *RawData) FromDir(d ...string) error {
	for _, rd := range d {
		// first, read properties
		data, err := grabFile(filepath.Join(rd, "_properties.json"))
		if err != nil {
			return err
		}

		var jsonDec map[string]string
		err = json.Unmarshal(data, &jsonDec)
		if err != nil {
			return err
		}

		// set data
		for k, v := range jsonDec {
			res.PageProperties[strings.Replace(k, "_", "-", -1)] = v
		}

		// now, let's grab all files (TODO read dir rather than relying on glob)
		m, err := filepath.Glob(filepath.Join(rd, "*.tpl"))
		if err != nil {
			return err
		}

		for _, n := range m {
			data, err = grabFile(n)
			if err != nil {
				return err
			}

			base := filepath.Base(n)

			res.TemplateData[base[:len(base)-4]] = string(data)
		}
	}
	return nil
}

func (res *RawData) FromVfs(src fs.FS, d ...string) error {
	for _, rd := range d {
		// first, read properties
		data, err := grabVfsFile(src, path.Join(rd, "_properties.json"))
		if err != nil {
			// not found?
			return nil
		}

		var jsonDec map[string]string
		err = json.Unmarshal(data, &jsonDec)
		if err != nil {
			return err
		}

		// set data
		for k, v := range jsonDec {
			res.PageProperties[strings.Replace(k, "_", "-", -1)] = v
		}

		// now, let's grab all files (TODO read dir rather than relying on glob)
		m, err := fs.ReadDir(src, rd)
		if err != nil {
			return err
		}

		for _, nfo := range m {
			n := nfo.Name()
			if !strings.HasSuffix(n, ".tpl") {
				continue
			}

			data, err = grabVfsFile(src, path.Join(rd, n))
			if err != nil {
				return err
			}

			res.TemplateData[n[:len(n)-4]] = string(data)
		}
	}
	return nil
}
