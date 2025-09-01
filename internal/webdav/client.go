package webdav

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	gowebdav "github.com/emersion/go-webdav"
)

type Client struct{ c *gowebdav.Client }

type Opts struct {
	BaseURL string
	Auth    string
	User    string
	Pass    string
	Token   string
}

type bearer struct {
	base gowebdav.HTTPClient
	tok  string
}

func (b bearer) Do(r *http.Request) (*http.Response, error) {
	if b.base == nil {
		b.base = http.DefaultClient
	}
	cl := r.Clone(r.Context())
	cl.Header.Set("Authorization", "Bearer "+b.tok)
	return b.base.Do(cl)
}

func New(o Opts) (*Client, error) {
	endpoint := strings.TrimSpace(o.BaseURL)
	if endpoint == "" {
		return nil, fmt.Errorf("empty base_url")
	}
	var hc gowebdav.HTTPClient
	switch strings.ToLower(strings.TrimSpace(o.Auth)) {
	case "", "basic":
		hc = gowebdav.HTTPClientWithBasicAuth(nil, o.User, o.Pass)
	case "bearer":
		hc = bearer{tok: o.Token}
	default:
		return nil, fmt.Errorf("unsupported auth")
	}
	c, err := gowebdav.NewClient(hc, endpoint)
	if err != nil {
		return nil, err
	}
	return &Client{c: c}, nil
}

func (c *Client) Mkdir(ctx context.Context, relDir string) error {
	relDir = strings.Trim(relDir, "/")
	if relDir == "" {
		return nil
	}
	cur := ""
	for _, seg := range strings.Split(relDir, "/") {
		if seg == "" {
			continue
		}
		cur = strings.Trim(strings.Join([]string{cur, seg}, "/"), "/")
		if err := c.c.Mkdir(ctx, cur); err != nil {
			if fi, e := c.c.Stat(ctx, cur); e == nil && fi.IsDir {
				continue
			}
			return err
		}
	}
	return nil
}

func (c *Client) Create(ctx context.Context, relPath string, data []byte) error {
	w, err := c.c.Create(ctx, strings.Trim(relPath, "/"))
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

func (c *Client) Exists(ctx context.Context, relPath string) (bool, error) {
	_, err := c.c.Stat(ctx, strings.Trim(relPath, "/"))
	if err != nil {
		return false, err
	}
	return true, nil
}

func Join(dir, file string) string {
	dir = strings.Trim(dir, "/")
	file = strings.Trim(file, "/")
	if dir == "" {
		return file
	}
	return dir + "/" + file
}

