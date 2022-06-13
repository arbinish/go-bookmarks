package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/arbinish/go-bookmarks/bookmarks"
)

type client struct {
	url     string
	timeout int
	client  *http.Client
}

func newClient(url string, timeout int) *client {
	return &client{
		url:     url,
		timeout: timeout,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

func (c *client) delete(name string) bool {
	if c.findByParam("name", name) == nil {
		fmt.Printf("%s: does not exist\n", name)
		return false
	}
	req, err := http.NewRequest(http.MethodDelete, c.url+"/api/v1/delete/"+name, nil)
	if err != nil {
		fmt.Println("unable to init request", err)
		return false
	}
	_, err = c.client.Do(req)
	return err == nil
}

func (c *client) dump() []*bookmarks.Bookmark {
	url := c.url + "/api/v1/dump"
	resp, err := c.client.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var b = make([]*bookmarks.Bookmark, 0)
	if err = dec.Decode(&b); err != nil {
		fmt.Println("decoding failed", err)
		return nil
	}
	return b
}

func (c *client) create(name, bookmarkURL, tags string) bool {
	_url := c.url + "/api/v1/create"
	var params = make(url.Values)
	params.Add("name", name)
	params.Add("tags", tags)
	params.Add("url", bookmarkURL)
	resp, err := c.client.PostForm(_url, params)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func (c *client) findByParam(param, value string) []*bookmarks.Bookmark {
	url := fmt.Sprintf("%s%s?%s=%s", c.url, "/api/v1/find", param, value)
	resp, err := c.client.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	dec := json.NewDecoder(resp.Body)
	var b = make([]*bookmarks.Bookmark, 0)
	if err = dec.Decode(&b); err != nil {
		fmt.Println("decoding failed", err)
		return nil
	}
	return b
}
