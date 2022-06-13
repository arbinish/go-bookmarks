package bookmarks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type application struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	db       *db
	sync     chan int
	numSaved int
}

func NewApp(info, err *log.Logger, d *db) *application {
	c := make(chan int, 1)
	return &application{
		infoLog:  info,
		errorLog: err,
		db:       d,
		sync:     c,
		numSaved: 0,
	}
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	enc := json.NewEncoder(w)
	if err := enc.Encode(app.db); err != nil {
		app.errorLog.Printf("encoding error: %s\n", err.Error())
		fmt.Fprintf(w, "%s", err.Error())
	}
}

func (app *application) getBookmarkByTag(w http.ResponseWriter, r *http.Request) {
	tag := strings.TrimPrefix(r.URL.Path, "/api/v1/tags/")
	if tag == "" {
		http.Error(w, "Missing Tag name", http.StatusBadRequest)
		return
	}
	result, ok := tagIndex[tag]
	if !ok {
		http.Error(w, fmt.Sprintf("%s: No such tag", tag), http.StatusNotFound)
		app.errorLog.Printf("%s: no such tag\n", tag)
		return
	}
	for _, r := range result {
		r.Update()
	}

	var buf bytes.Buffer

	mw := io.MultiWriter(w, &buf)
	enc := json.NewEncoder(mw)
	if err := enc.Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.infoLog.Printf("%s %s [size=%d]\n", r.Method, r.URL.Path, len(buf.String()))
}

func (app *application) getTags(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	response := make([]string, 0)

	for t := range tagIndex {
		response = append(response, t)
	}
	mw := io.MultiWriter(w, &buf)
	enc := json.NewEncoder(mw)
	if err := enc.Encode(response); err != nil {
		app.errorLog.Printf("encoding error: %s\n", err.Error())
		fmt.Fprintf(w, "%s", err.Error())
		return
	}
	app.infoLog.Printf("%s %s [size=%d]\n", r.Method, r.URL.Path, len(buf.String()))
}

func (app *application) find(w http.ResponseWriter, r *http.Request) {
	var index = make(map[string]struct{})
	name := r.URL.Query().Get("name")
	enc := json.NewEncoder(w)
	var ok, valid bool
	var q *Bookmark
	if name != "" {
		if _, ok = nameIndex[name]; !ok {
			http.Error(w, fmt.Sprintf("%s: not found", name), http.StatusNotFound)
			return
		}
		valid = true
		index[name] = struct{}{}
	}
	url := r.URL.Query().Get("url")
	if url != "" {
		if q, ok = urlIndex[url]; !ok {
			http.Error(w, fmt.Sprintf("%s: not found", url), http.StatusNotFound)
			return
		}
		index[q.Name] = struct{}{}
		valid = true
	}
	var tags []string
	var tagMap = make(map[string]bool)

	tag := r.URL.Query().Get("tag")
	if tag != "" {
		tags = strings.Split(tag, ",")
		for _, _tag := range tags {
			if _, ok = tagIndex[_tag]; !ok {
				http.Error(w, fmt.Sprintf("%s: not found", _tag), http.StatusNotFound)
				return
			}
			for _, e := range tagIndex[_tag] {
				index[e.Name] = struct{}{}
				tagMap[_tag] = true
			}
		}
		valid = true
	}
	if !valid {
		http.Error(w, "One of url, tag, name param missing", http.StatusBadRequest)
	} else {
		var r = make([]*Bookmark, 0)
		for b := range index {
			for _, t := range nameIndex[b].Tags {
				if _, ok := tagMap[t]; ok {
					nameIndex[b].Update()
					r = append(r, nameIndex[b])
				}

			}
		}
		enc.Encode(r)
	}
}

func (app *application) createBookmark(w http.ResponseWriter, r *http.Request) {
	paramsExpected := []string{"name", "url", "tags"}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "incorrect method", http.StatusMethodNotAllowed)
		return
	}
	var name, url string
	var tags = make([]string, 0)
	for _, param := range paramsExpected {
		if r.FormValue(param) == "" {
			http.Error(w, fmt.Sprintf("missing param %s", param), http.StatusBadRequest)
			return
		}
	}
	name = r.FormValue("name")
	url = r.FormValue("url")
	for _, s := range strings.Split(r.FormValue("tags"), ",") {
		tags = append(tags, strings.TrimSpace(s))
	}
	bk := NewBookmark(name, url, tags)
	if err := app.db.Add(bk); err != nil {
		app.errorLog.Printf("failed to create bookmark %s: %s\n", bk, err)
		fmt.Fprintf(w, "%s", err)
	} else {
		app.infoLog.Printf("created: %s", bk)
		fmt.Fprintf(w, "created %s", bk.Name)
		app.Save()
	}
}

func jsonMiddleware(log *log.Logger, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "application/json")
		next(w, r)
	})
}

func (app *application) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/api/v1/tags", jsonMiddleware(app.infoLog, app.getTags))
	mux.HandleFunc("/api/v1/tags/", jsonMiddleware(app.infoLog, app.getBookmarkByTag))
	mux.HandleFunc("/api/v1/find", jsonMiddleware(app.infoLog, app.find))
	mux.HandleFunc("/api/v1/create", app.createBookmark)
	mux.HandleFunc("/api/v1/save", app.Sync)
	mux.HandleFunc("/api/v1/dump", app.Dump)
	mux.HandleFunc("/api/v1/delete/", jsonMiddleware(app.infoLog, app.Delete))
	return mux
}

func (app *application) Dump(w http.ResponseWriter, r *http.Request) {
	b := app.db.Dump()
	enc := json.NewEncoder(w)
	if err := enc.Encode(b); err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}
}

func (app *application) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "invalid method", http.StatusBadRequest)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/api/v1/delete/")
	if name == "" {
		http.Error(w, "missing name", http.StatusBadRequest)
		return
	}
	if app.db.DeleteBookmark(name) == nil {
		fmt.Fprintf(w, "%s deleted", name)
		// persist data immediately
		app.Save()
	} else {
		app.errorLog.Printf("missing bookmark by name [%s]\n", name)
		fmt.Fprintf(w, "missing bookmark by name %s", name)
	}
}

func (app *application) Save() error {
	app.infoLog.Println("acquiring lock")
	app.sync <- 1
	app.infoLog.Println("got lock")
	app.numSaved++
	file, err := os.Create("db.dump")
	stat, _ := os.Create("db.dump.stat")
	if err != nil {
		app.errorLog.Println(err)
		return err
	}
	defer func() {
		file.Close()
		stat.Close()
		<-app.sync
		app.infoLog.Println("released lock")
	}()
	enc := json.NewEncoder(file)
	err = enc.Encode(app.db)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}
	stat.WriteString(fmt.Sprintf(
		"last_saved=%d\nsave_count=%d\nsize=%d\n",
		time.Now().Unix(), app.numSaved, app.db.Size()))
	return nil
}

func (app *application) Load() {
	file, err := os.Open("db.dump")
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	err = dec.Decode(app.db)
	if err != nil {
		app.errorLog.Println("failed to decode data from persistent store", err)
		return
	}
	app.infoLog.Println("successfully loaded data from persistent store")
	// Update various indices
	app.db.updateIndex()
	app.infoLog.Println("successfully updated indices")
}

func (app *application) Sync(w http.ResponseWriter, r *http.Request) {
	if app.Save() != nil {
		fmt.Fprintf(w, "failed to persist data\n")
	} else {
		fmt.Fprintf(w, "data persisted successfully\n")
	}
}
