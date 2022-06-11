package bookmarks

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type bookmark struct {
	Name     string
	Tags     []string
	Url      string
	Created  int64
	Accessed int64
	Views    int32
}

func (b bookmark) String() string {
	return fmt.Sprintf("%s | %s | %s | Views: %d", b.Name, b.Url, strings.Join(b.Tags, ","), b.Views)
}

type db []*bookmark

var tagIndex = make(map[string][]*bookmark, 0)
var urlIndex = make(map[string]*bookmark)
var nameIndex = make(map[string]*bookmark)

func NewDB() db {
	return make([]*bookmark, 0)
}

func (d db) FindUrl(url string) *bookmark {
	if b, ok := urlIndex[url]; ok {
		return b
	}
	return nil
}

func (d db) Find(name string) (int, *bookmark) {
	for i, k := range d {
		if k.Name == name {
			k.Views++
			return i, k
		}
	}
	return -1, nil
}

func (d db) FindbyTags(tags ...string) []*bookmark {
	r := make([]*bookmark, 0)
	for _, tag := range tags {
		if bookmarkList, ok := tagIndex[tag]; ok {
			for _, b := range bookmarkList {
				b.Views++
				r = append(r, b)
			}
		}
	}
	return r
}

func (d *db) updateIndex() {
	for _, b := range *d {
		nameIndex[b.Name] = b
		urlIndex[b.Url] = b
		for _, t := range b.Tags {
			tagIndex[t] = append(tagIndex[t], b)
		}
	}
}

func (d *db) DeleteBookmark(name string) error {
	i, b := d.Find(name)
	if b == nil {
		return errors.New(name + ": no such record")
	}
	(*d)[i] = (*d)[0]
	*d = (*d)[1:]
	delete(nameIndex, b.Name)
	delete(urlIndex, b.Url)
	for _, t := range b.Tags {
		b := tagIndex[t]
		if len(b) == 1 {
			delete(tagIndex, t)
			continue
		}
		for i, entry := range b {
			if entry.Name == name {
				tagIndex[t][i] = tagIndex[t][0]
				tagIndex[t] = tagIndex[t][1:]
			}
		}
	}
	return nil
}

func (d db) Dump() {
	for _, k := range d {
		fmt.Println("\t", k)
	}
	fmt.Println("dumping urls")
	for u, v := range urlIndex {
		fmt.Println("Url", u, "value", v)
	}
}

func (d db) Size() int {
	return len(d)
}

func (d *db) Add(b *bookmark) error {
	if _, found := d.Find(b.Name); found != nil {
		return errors.New("[SKIP] entry " + b.Name + " already exists.")
	}
	*d = append(*d, b)
	for _, tag := range b.Tags {
		tagIndex[tag] = append(tagIndex[tag], b)
	}
	urlIndex[b.Url] = b
	nameIndex[b.Name] = b
	return nil
}

func (b *bookmark) Update() error {
	b.Views++
	b.Accessed = time.Now().Unix()
	return nil
}

// NewBookmark returns a new bookmark record
func NewBookmark(name, url string, tags []string) *bookmark {
	t := make([]string, len(tags))
	copy(t, tags)
	return &bookmark{
		Name:     name,
		Url:      url,
		Tags:     t,
		Created:  time.Now().Unix(),
		Accessed: 0,
		Views:    0,
	}
}
