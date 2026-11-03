package panel

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"net/http"
	"strings"
	"sun-panel/lib/iconcache"
	"sun-panel/models"
	"sun-panel/models/datatype"
	"sync"
	"sync/atomic"
	"testing"
)

func TestSavedCardsGetMissingDetailsDiscovered(t *testing.T) {
	db := testDB(t)
	group := createGroup(t, db, 1, 0)
	var down atomic.Int32
	oldCache := iconcache.Default
	iconcache.Default = iconcache.New(t.TempDir(), &http.Client{Transport: iconTransport(func(r *http.Request) (*http.Response, error) {
		body := []byte(`<html><head><title>Found Name · Tagline</title><meta property="og:site_name" content="Found Name"><meta name="description" content="Found description"><link rel="apple-touch-icon" href="/touch.png"></head></html>`)
		switch {
		case r.URL.Host == "down.example":
			down.Add(1)
			return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}, Request: r}, nil
		case r.URL.Path == "/touch.png":
			var data bytes.Buffer
			png.Encode(&data, image.NewNRGBA(image.Rect(0, 0, 4, 4)))
			body = data.Bytes()
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(body)), Header: http.Header{}, Request: r}, nil
	})})
	t.Cleanup(func() { iconFill.wg.Wait(); iconcache.Default = oldCache })
	icon := func(url string) datatype.ItemIconIconInfo {
		var card models.ItemIcon
		db.Where("url=?", url).First(&card)
		json.Unmarshal([]byte(card.IconJson), &card.Icon)
		return card.Icon
	}

	res := request(t, 1, false, (&ItemIcon{}).Edit, map[string]interface{}{"url": "https://site.example", "title": "site", "itemIconGroupId": group.ID})
	if res["code"] != float64(0) {
		t.Fatal(res)
	}
	res = request(t, 1, false, (&ItemIcon{}).AddMultiple, []map[string]interface{}{
		{"url": "https://imported.example", "title": "imported.example", "itemIconGroupId": group.ID},
		{"url": "https://text.example", "title": "text", "itemIconGroupId": group.ID, "icon": map[string]interface{}{"itemType": 1, "text": "T"}},
		{"url": "https://down.example", "title": "down", "itemIconGroupId": group.ID},
	})
	if res["code"] != float64(0) {
		t.Fatal(res)
	}
	iconFill.wg.Wait()
	for _, url := range []string{"https://site.example", "https://imported.example"} {
		if got := icon(url); got.ItemType != 2 || got.Src != url+"/touch.png" {
			t.Fatal("icon not saved", url, got)
		}
	}
	// Address-only names and empty descriptions are filled; chosen names stay.
	var site, imported models.ItemIcon
	db.Where("url=?", "https://site.example").First(&site)
	db.Where("url=?", "https://imported.example").First(&imported)
	if site.Title != "site" || site.Description != "Found description" || imported.Title != "Found Name" || imported.Description != "Found description" {
		t.Fatal(site.Title, site.Description, imported.Title, imported.Description)
	}
	if got := icon("https://text.example"); got.ItemType != 1 || got.Src != "" {
		t.Fatal("existing icon replaced", got)
	}
	if got := icon("https://down.example"); got.Src != "" {
		t.Fatal(got)
	}
	// An unreachable site is not queried again on every save.
	before := down.Load()
	if before == 0 {
		t.Fatal("unreachable site not queried")
	}
	request(t, 1, false, (&ItemIcon{}).Edit, map[string]interface{}{"url": "https://down.example", "title": "down", "itemIconGroupId": group.ID})
	iconFill.wg.Wait()
	if down.Load() != before {
		t.Fatal("failed site retried immediately")
	}
}

func TestRefillReplacesDeadAndOnlineIconsOnly(t *testing.T) {
	db := testDB(t)
	group := createGroup(t, db, 1, 0)
	var pages sync.Map
	oldCache := iconcache.Default
	iconcache.Default = iconcache.New(t.TempDir(), &http.Client{Transport: iconTransport(func(r *http.Request) (*http.Response, error) {
		body := []byte(`<html><head><title>Site</title><meta name="description" content="Found description"><link rel="apple-touch-icon" href="/touch.png"></head></html>`)
		switch {
		case r.URL.Path == "/dead.png":
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}, Request: r}, nil
		case strings.HasSuffix(r.URL.Path, ".png"):
			var data bytes.Buffer
			png.Encode(&data, image.NewNRGBA(image.Rect(0, 0, 4, 4)))
			body = data.Bytes()
		default:
			pages.Store(r.URL.Host, true)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(body)), Header: http.Header{}, Request: r}, nil
	})})
	t.Cleanup(func() { iconFill.wg.Wait(); iconcache.Default = oldCache })
	add := func(site string, icon datatype.ItemIconIconInfo) {
		raw, _ := json.Marshal(icon)
		db.Create(&models.ItemIcon{UserId: 1, ItemIconGroupId: int(group.ID), Url: site, Title: "chosen", Description: "kept", IconJson: string(raw)})
	}
	add("https://online.example", datatype.ItemIconIconInfo{ItemType: 3, Text: "mdi:github"})
	add("https://dead.example", datatype.ItemIconIconInfo{ItemType: 2, Src: "https://cdn.example/dead.png"})
	add("https://alive.example", datatype.ItemIconIconInfo{ItemType: 2, Src: "https://cdn.example/alive.png"})
	add("https://text.example", datatype.ItemIconIconInfo{ItemType: 1, Text: "T"})
	res := request(t, 1, false, (&ItemIcon{}).Refill, nil)
	if res["code"] != float64(0) || res["data"].(map[string]interface{})["queued"] != float64(3) {
		t.Fatal(res)
	}
	iconFill.wg.Wait()
	icon := func(site string) datatype.ItemIconIconInfo {
		var card models.ItemIcon
		db.Where("url=?", site).First(&card)
		json.Unmarshal([]byte(card.IconJson), &card.Icon)
		if card.Title != "chosen" || card.Description != "kept" {
			t.Fatal("chosen details replaced", card.Title, card.Description)
		}
		return card.Icon
	}
	for _, site := range []string{"https://online.example", "https://dead.example"} {
		if got := icon(site); got.ItemType != 2 || got.Src != site+"/touch.png" {
			t.Fatal("icon not replaced", site, got)
		}
	}
	if got := icon("https://alive.example"); got.Src != "https://cdn.example/alive.png" {
		t.Fatal("working icon replaced", got)
	}
	if got := icon("https://text.example"); got.ItemType != 1 {
		t.Fatal("text icon replaced", got)
	}
	if _, read := pages.Load("alive.example"); read {
		t.Fatal("healthy card's site was queried")
	}
}
