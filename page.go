package gen

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"path"
	"path/filepath"
)

// Assets corresponding to all files in the dir directory.
//
// Returns an error if there were any problems reading the directory or files.
func Assets(dir string) ([]Page, error) {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var as []*Asset
	for _, f := range fs {
		a, err := NewAsset(filepath.Join(dir, f.Name()))
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}
	return as, nil
}

// Asset is a static file.
type Asset struct {
	Page
}

// NewAsset from static file at path p that will be accessed from only its file
// name.
//
// Returns an error if the file couldn't be read.
func NewAsset(p string) (*Asset, error) {
	bs, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	return &Asset{NewPage(path.Base(p), bytes.NewBuffer(bs))}, nil
}

// Sitemap is an XML file describing the website's structure.
type Sitemap struct {
	Page
}

// NewSitemap for the website at url u made of all ps Pages.
//
// The Sitemap will filter out any Page which isn't an HTML page automatically.
func NewSitemap(u string, ps []Page) (*Sitemap, error) {
	sm := sitemap{}
	sm.Version = "http://www.sitemaps.org/schemas/sitemap/0.9"
	for _, p := range filter(".html", ps) {
		if p.Path() == "index.html" {
			sm.URLs = append(sm.URLs, url{Loc: u})
		} else if filepath.Base(p.Path()) == "index.html" {
			path := filepath.Dir(p.Path()) + "/"
			sm.URLs = append(sm.URLs, url{Loc: u + path})
		} else {
			sm.URLs = append(sm.URLs, url{Loc: u + "/" + p.Path()})
		}
	}
	buf := bytes.NewBufferString(xml.Header)
	enc := xml.NewEncoder(buf)
	enc.Indent("", "    ")
	if err := enc.Encode(sm); err != nil {
		return nil, err
	}
	return &Sitemap{NewPage("sitemap.xml", buf)}, nil
}

// sitemap is an XML struct which defines what a sitemap looks like in XML.
type sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	Version string   `xml:"xmlns,attr"`
	URLs    []url    `xml:"url"`
}

// url is an XML struct which defines what a URL looks like in XML.
type url struct {
	Loc string `xml:"loc"`
}

// filter Pages based on their extensions.
func filter(ext string, ps []Page) []Page {
	var fps []Page
	for _, p := range ps {
		if filepath.Ext(p.Path()) == ext {
			fps = append(fps, p)
		}
	}
	return fps
}
