package panel

import (
	"bytes"
	"context"
	"encoding/json"
	"gorm.io/gorm"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sun-panel/global"
	"sun-panel/lib/iconcache"
	"sun-panel/lib/iniConfig"
	"sun-panel/models"
	"sun-panel/models/datatype"
	"sync"
	"testing"
)

type iconTransport func(*http.Request) (*http.Response, error)

func (f iconTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCachedFileLifecycleAcrossReplacementAndGroups(t *testing.T) {
	db := testDB(t)
	group := createGroup(t, db, 1, 0)
	other := createGroup(t, db, 2, 0)
	createGroup(t, db, 2, 0)
	oldCache := iconcache.Default
	iconcache.Default = iconcache.New(t.TempDir(), &http.Client{Transport: iconTransport(func(r *http.Request) (*http.Response, error) {
		img := image.NewNRGBA(image.Rect(0, 0, 8, 8))
		shade := uint8(50)
		if strings.Contains(r.URL.Path, "new") {
			shade = 200
		}
		img.SetNRGBA(0, 0, color.NRGBA{shade, 0, 0, 255})
		var data bytes.Buffer
		png.Encode(&data, img)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(data.Bytes())), Header: http.Header{}, Request: r}, nil
	})})
	t.Cleanup(func() { iconcache.Default = oldCache })
	a, b := "https://icons.example/old.png", "https://icons.example/new.png"
	local, _ := iconcache.Default.Fetch(context.Background(), a)
	next, _ := iconcache.Default.Fetch(context.Background(), b)
	save := func(user uint, g uint, site, source string) map[string]interface{} {
		res := request(t, user, false, (&ItemIcon{}).Edit, map[string]interface{}{"url": site, "title": "site", "itemIconGroupId": g, "icon": map[string]interface{}{"itemType": 2, "src": source, "cachedSrc": "/untrusted"}})
		if res["code"] != float64(0) {
			t.Fatal(res)
		}
		return res
	}
	save(1, group.ID, "https://example.com", a)
	save(2, other.ID, "https://other.example", a)
	save(1, group.ID, "https://example.com", b)
	asset := func(url string) string {
		return filepath.Join(iconcache.Default.Root, "assets", strings.TrimPrefix(url, "/icon-cache/"))
	}
	if _, err := os.Stat(asset(local)); err != nil {
		t.Fatal("shared cached file removed", err)
	}
	var card models.ItemIcon
	db.Where("user_id=?", 1).First(&card)
	if strings.Contains(card.IconJson, "cachedSrc") {
		t.Fatal("derived icon source persisted")
	}
	res := request(t, 2, false, (&ItemIconGroup{}).Deletes, map[string]interface{}{"ids": []uint{other.ID}})
	if res["code"] != float64(0) {
		t.Fatal(res)
	}
	if _, err := os.Stat(asset(local)); !os.IsNotExist(err) {
		t.Fatal("orphan cached file retained", err)
	}
	if _, err := os.Stat(asset(next)); err != nil {
		t.Fatal("current cached icon removed", err)
	}
}

func TestConcurrentSameURLCannotCreateDuplicates(t *testing.T) {
	db := testDB(t)
	group := createGroup(t, db, 1, 0)
	var wg sync.WaitGroup
	failures := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cardMu.Lock()
			defer cardMu.Unlock()
			card := models.ItemIcon{UserId: 1, ItemIconGroupId: int(group.ID), Url: "https://example.com/", Title: "latest"}
			old := []models.ItemIcon{}
			err := db.Transaction(func(tx *gorm.DB) error { return saveCard(tx, &card, 1, &old) })
			if err != nil {
				failures <- err
			}
		}()
	}
	wg.Wait()
	close(failures)
	for err := range failures {
		t.Fatal(err)
	}
	var count int64
	db.Model(&models.ItemIcon{}).Count(&count)
	if count != 1 {
		t.Fatal("concurrent duplicate rows", count)
	}
}

func TestSameURLReplacesLatestAndReaddAfterDelete(t *testing.T) {
	db := testDB(t)
	group := createGroup(t, db, 1, 0)
	api := &ItemIcon{}
	save := func(raw, title string) map[string]interface{} {
		return request(t, 1, false, api.Edit, map[string]interface{}{"url": raw, "title": title, "itemIconGroupId": group.ID})
	}
	first := save("https://EXAMPLE.com:443/", "old")
	if first["code"] != float64(0) {
		t.Fatal(first)
	}
	id := first["data"].(map[string]interface{})["id"]
	second := save("https://example.com", "new")
	if second["code"] != float64(0) || second["data"].(map[string]interface{})["id"] != id {
		t.Fatal(second)
	}
	var cards []models.ItemIcon
	db.Find(&cards)
	if len(cards) != 1 || cards[0].Title != "new" {
		t.Fatal(cards)
	}
	// Meaningful query differences stay distinct, and another user owns their own copy.
	if res := save("https://example.com/?app=2", "query"); res["code"] != float64(0) {
		t.Fatal(res)
	}
	other := createGroup(t, db, 2, 0)
	if res := request(t, 2, false, api.Edit, map[string]interface{}{"url": "https://example.com", "title": "other", "itemIconGroupId": other.ID}); res["code"] != float64(0) {
		t.Fatal(res)
	}
	if res := request(t, 1, false, api.Deletes, map[string]interface{}{"ids": []interface{}{id}}); res["code"] != float64(0) {
		t.Fatal(res)
	}
	if res := save("https://example.com/", "restored"); res["code"] != float64(0) {
		t.Fatal("soft-deleted unique key blocked readd", res)
	}
	db.Find(&cards)
	if len(cards) != 3 {
		t.Fatal("incorrect card count", len(cards))
	}
}

func TestLegacyDuplicatesAndBulkLastWins(t *testing.T) {
	db := testDB(t)
	group := createGroup(t, db, 1, 0)
	for _, raw := range []string{"https://example.com/", "https://EXAMPLE.com:443/"} {
		db.Create(&models.ItemIcon{UserId: 1, ItemIconGroupId: int(group.ID), Url: raw, Title: "legacy"})
	}
	res := request(t, 1, false, (&ItemIcon{}).AddMultiple, []map[string]interface{}{{"url": "https://example.com", "title": "first", "itemIconGroupId": group.ID}, {"url": "https://example.com/", "title": "latest", "itemIconGroupId": group.ID}})
	if res["code"] != float64(0) {
		t.Fatal(res)
	}
	var cards []models.ItemIcon
	db.Find(&cards)
	if len(cards) != 1 || cards[0].Title != "latest" || cards[0].URLKey == nil {
		t.Fatal(cards)
	}
}

func TestBatchedReadRespectsVisibilityAndOldClients(t *testing.T) {
	db := testDB(t)
	public := createGroup(t, db, 1, 0)
	private := createGroup(t, db, 1, 1)
	db.Create(&models.ItemIcon{UserId: 1, ItemIconGroupId: int(public.ID), Title: "visible"})
	db.Create(&models.ItemIcon{UserId: 1, ItemIconGroupId: int(private.ID), Title: "hidden"})
	list := responseList(t, request(t, 1, true, (&ItemIconGroup{}).GetList, map[string]interface{}{"includeItems": true}))
	if len(list) != 1 {
		t.Fatal(list)
	}
	items := list[0].(map[string]interface{})["items"].([]interface{})
	if len(items) != 1 || items[0].(map[string]interface{})["title"] != "visible" {
		t.Fatal(items)
	}
	old := responseList(t, request(t, 1, true, (&ItemIconGroup{}).GetList, nil))
	if _, exists := old[0].(map[string]interface{})["items"]; exists {
		t.Fatal("old API contract changed")
	}
}

func TestDeletingSharedUploadRemovesLastReferenceOnly(t *testing.T) {
	db := testDB(t)
	group := createGroup(t, db, 1, 0)
	db.AutoMigrate(&models.File{}, &models.UserConfig{})
	// Use a relative upload root as in the deployed configuration.
	configPath := filepath.Join(t.TempDir(), "conf.ini")
	os.WriteFile(configPath, []byte("[base]\nsource_path=./uploads\n"), 0600)
	oldConfig := global.Config
	global.Config = iniConfig.NewIniConfig(configPath)
	t.Cleanup(func() { global.Config = oldConfig })
	uploadRoot := filepath.Join(t.TempDir(), "uploads")
	os.MkdirAll(uploadRoot, 0700)
	global.Config.Config.Section("base").Key("source_path").SetValue(uploadRoot)
	publicPath := "/" + filepath.ToSlash(uploadRoot)[1:] + "/shared.png"
	file := filepath.Join(uploadRoot, "shared.png")
	os.WriteFile(file, []byte("icon"), 0600)
	icon, _ := json.Marshal(datatype.ItemIconIconInfo{ItemType: 2, Src: publicPath})
	a := models.ItemIcon{UserId: 1, ItemIconGroupId: int(group.ID), Title: "a", IconJson: string(icon)}
	b := a
	db.Create(&a)
	db.Create(&b)
	oldCache := iconcache.Default
	iconcache.Default = nil
	t.Cleanup(func() { iconcache.Default = oldCache })
	for i, id := range []uint{a.ID, b.ID} {
		res := request(t, 1, false, (&ItemIcon{}).Deletes, map[string]interface{}{"ids": []uint{id}})
		if res["code"] != float64(0) {
			t.Fatal(res)
		}
		_, err := os.Stat(file)
		if i == 0 && err != nil {
			t.Fatal("shared file removed")
		}
		if i == 1 && !os.IsNotExist(err) {
			t.Fatal("orphan file retained", err)
		}
	}
}
