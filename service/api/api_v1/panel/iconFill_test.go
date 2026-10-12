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
	"sync/atomic"
	"testing"
)

func TestSavedCardsWithoutIconGetOneDiscovered(t *testing.T) {
	db := testDB(t)
	group := createGroup(t, db, 1, 0)
	var down atomic.Int32
	oldCache := iconcache.Default
	iconcache.Default = iconcache.New(t.TempDir(), &http.Client{Transport: iconTransport(func(r *http.Request) (*http.Response, error) {
		body := []byte(`<html><head><link rel="apple-touch-icon" href="/touch.png"></head></html>`)
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
		{"url": "https://imported.example", "title": "imported", "itemIconGroupId": group.ID},
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
