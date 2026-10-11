package iconcache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sun-panel/lib/remoteimage"
	"sync"
	"time"
)

var Default *Cache

// Coordinates bookmark reference mutations with scheduled collection.
var ReferenceMu sync.Mutex
var assetName = regexp.MustCompile(`^[0-9a-f]{64}\.(png|svg|ico)$`)
var assetPath = regexp.MustCompile(`/icon-cache/[0-9a-f]{64}\.(?:png|svg|ico)`)

type Entry struct {
	File    string    `json:"file"`
	Created time.Time `json:"created"`
}
type job struct {
	done      chan struct{}
	entry     Entry
	err       error
	discarded bool
}
type Cache struct {
	Root     string
	Client   *http.Client
	mu       sync.Mutex
	jobs     map[string]*job
	failures map[string]time.Time
	slots    chan struct{}
}

func New(root string, client *http.Client) *Cache {
	if client == nil {
		client = remoteimage.NewClient()
	}
	return &Cache{Root: root, Client: client, jobs: map[string]*job{}, failures: map[string]time.Time{}, slots: make(chan struct{}, 4)}
}

func key(source string) string {
	sum := sha256.Sum256([]byte("192-v1\x00" + strings.TrimSpace(source)))
	return hex.EncodeToString(sum[:])
}

func (c *Cache) readEntry(k string) Entry {
	var entry Entry
	data, err := os.ReadFile(filepath.Join(c.Root, "index", k+".json"))
	if err != nil || json.Unmarshal(data, &entry) != nil || !assetName.MatchString(entry.File) {
		return Entry{}
	}
	if info, err := os.Stat(filepath.Join(c.Root, "assets", entry.File)); err != nil || !info.Mode().IsRegular() {
		return Entry{}
	}
	return entry
}

// Resolve never waits on a remote host. Existing source URLs remain in the database.
func (c *Cache) Resolve(source string) string {
	if c == nil {
		return ""
	}
	if _, err := remoteimage.ValidateURL(source); err != nil {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	k := key(source)
	if entry := c.readEntry(k); entry.File != "" {
		return "/icon-cache/" + entry.File
	}
	c.startLocked(source, k)
	return ""
}

func (c *Cache) startLocked(source, k string) *job {
	if pending := c.jobs[k]; pending != nil {
		return pending
	}
	if until := c.failures[k]; time.Now().Before(until) {
		return nil
	}
	if len(c.jobs) >= 256 {
		return nil
	}
	pending := &job{done: make(chan struct{})}
	c.jobs[k] = pending
	go func() {
		c.slots <- struct{}{}
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		data, _, _, err := remoteimage.Get(ctx, c.Client, source, 2<<20)
		cancel()
		var output []byte
		var extension string
		if err == nil {
			output, extension, err = prepare(data)
		}
		<-c.slots
		c.mu.Lock()
		defer c.mu.Unlock()
		defer func() { delete(c.jobs, k); close(pending.done) }()
		if pending.discarded {
			pending.err = errors.New("icon is no longer referenced")
			return
		}
		if err == nil {
			sum := sha256.Sum256(output)
			file := hex.EncodeToString(sum[:]) + extension
			err = os.MkdirAll(filepath.Join(c.Root, "assets"), 0700)
			if err == nil {
				err = os.MkdirAll(filepath.Join(c.Root, "index"), 0700)
			}
			if err == nil {
				err = writeAtomic(filepath.Join(c.Root, "assets", file), output)
			}
			if err == nil {
				pending.entry = Entry{File: file, Created: time.Now().UTC()}
				metadata, _ := json.Marshal(pending.entry)
				err = writeAtomic(filepath.Join(c.Root, "index", k+".json"), metadata)
			}
		}
		pending.err = err
		if err != nil {
			c.failures[k] = time.Now().Add(15 * time.Minute)
		}
	}()
	return pending
}

func (c *Cache) Fetch(ctx context.Context, source string) (string, error) {
	if c == nil {
		return "", errors.New("icon cache is unavailable")
	}
	if _, err := remoteimage.ValidateURL(source); err != nil {
		return "", err
	}
	c.mu.Lock()
	k := key(source)
	if entry := c.readEntry(k); entry.File != "" {
		c.mu.Unlock()
		return "/icon-cache/" + entry.File, nil
	}
	pending := c.startLocked(source, k)
	c.mu.Unlock()
	if pending == nil {
		return "", errors.New("icon is temporarily unavailable")
	}
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-pending.done:
		if pending.err != nil {
			return "", pending.err
		}
		return "/icon-cache/" + pending.entry.File, nil
	}
}

func writeAtomic(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".pending-")
	if err != nil {
		return err
	}
	name := file.Name()
	defer os.Remove(name)
	if _, err = file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

// Remove only old candidates, retaining shared files and fresh, unsaved previews.
func (c *Cache) Prune(sources, paths, candidates []string) error {
	if c == nil {
		return nil
	}
	keepKeys := map[string]bool{}
	keepFiles := map[string]bool{}
	removeKeys := map[string]bool{}
	removeFiles := map[string]bool{}
	for _, s := range sources {
		keepKeys[key(s)] = true
	}
	for _, p := range paths {
		for _, match := range assetPath.FindAllString(p, -1) {
			keepFiles[strings.TrimPrefix(match, "/icon-cache/")] = true
		}
	}
	for _, s := range candidates {
		if !keepKeys[key(s)] {
			removeKeys[key(s)] = true
		}
		if strings.HasPrefix(s, "/icon-cache/") {
			removeFiles[strings.TrimPrefix(s, "/icon-cache/")] = true
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range removeKeys {
		if pending := c.jobs[k]; pending != nil {
			pending.discarded = true
		}
		delete(c.failures, k)
	}
	entries, err := os.ReadDir(filepath.Join(c.Root, "index"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, file := range entries {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		k := strings.TrimSuffix(file.Name(), ".json")
		entry := c.readEntry(k)
		if entry.File == "" {
			continue
		}
		if !keepKeys[k] && (removeKeys[k] || (removeFiles[entry.File] && !keepFiles[entry.File]) || time.Since(entry.Created) > 24*time.Hour) {
			if err := os.Remove(filepath.Join(c.Root, "index", file.Name())); err != nil && !os.IsNotExist(err) {
				return err
			}
		} else {
			keepFiles[entry.File] = true
		}
	}
	files, err := os.ReadDir(filepath.Join(c.Root, "assets"))
	if err != nil {
		return err
	}
	for _, file := range files {
		if assetName.MatchString(file.Name()) && !keepFiles[file.Name()] {
			if err := os.Remove(filepath.Join(c.Root, "assets", file.Name())); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}

func (c *Cache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/icon-cache/")
	if !assetName.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	file := filepath.Join(c.Root, "assets", name)
	if info, err := os.Stat(file); err != nil || !info.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}
	contentType := map[string]string{".png": "image/png", ".svg": "image/svg+xml", ".ico": "image/x-icon"}[filepath.Ext(name)]
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'; style-src 'unsafe-inline'")
	w.Header().Set("ETag", `"`+name+`"`)
	if r.Header.Get("If-None-Match") == `"`+name+`"` {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	http.ServeFile(w, r, file)
}
