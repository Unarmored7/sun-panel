package panel

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"net/http/httptest"
	"path/filepath"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/models"
	"testing"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.ItemIconGroup{}, &models.ItemIcon{}); err != nil {
		t.Fatal(err)
	}
	old := global.Db
	global.Db = db
	sqlDB, _ := db.DB()
	t.Cleanup(func() { global.Db = old; sqlDB.Close() })
	return db
}

func request(t *testing.T, user uint, public bool, handler gin.HandlerFunc, body interface{}) map[string]interface{} {
	t.Helper()
	data, _ := json.Marshal(body)
	r := gin.New()
	r.POST("/", func(c *gin.Context) {
		c.Set("userInfo", models.User{BaseModel: models.BaseModel{ID: user}})
		if public {
			c.Set(base.GIN_GET_VISIT_MODE, base.VISIT_MODE_PUBLIC)
		}
		handler(c)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Response: %s", w.Body.String())
	}
	return response
}

func createGroup(t *testing.T, db *gorm.DB, user uint, hidden int) models.ItemIconGroup {
	t.Helper()
	g := models.ItemIconGroup{UserId: user, Title: "group", PublicVisitHide: hidden, CardStyle: `{"style":1,"textColor":"#ffffff"}`}
	if err := db.Create(&g).Error; err != nil {
		t.Fatal(err)
	}
	return g
}

func responseList(t *testing.T, res map[string]interface{}) []interface{} {
	t.Helper()
	if res["code"] != float64(0) {
		t.Fatalf("Request failed: %v", res)
	}
	return res["data"].(map[string]interface{})["list"].([]interface{})
}

func TestVisitorCannotReadHiddenGroupsOrCards(t *testing.T) {
	db := testDB(t)
	private := createGroup(t, db, 1, 1)
	public := createGroup(t, db, 1, 0)
	createGroup(t, db, 2, 0)
	db.Create(&models.ItemIcon{UserId: 1, ItemIconGroupId: int(private.ID), Title: "private"})
	db.Create(&models.ItemIcon{UserId: 1, ItemIconGroupId: int(public.ID), Title: "public"})
	groups := &ItemIconGroup{}
	items := &ItemIcon{}
	if list := responseList(t, request(t, 1, true, groups.GetList, nil)); len(list) != 1 {
		t.Fatalf("Visitor sees %d groups", len(list))
	}
	if list := responseList(t, request(t, 1, false, groups.GetList, nil)); len(list) != 2 {
		t.Fatalf("Owner sees %d groups", len(list))
	}
	if res := request(t, 1, true, items.GetListByGroupId, map[string]interface{}{"itemIconGroupId": private.ID}); res["code"] != float64(1005) {
		t.Fatal(res)
	}
	if list := responseList(t, request(t, 1, true, items.GetListByGroupId, map[string]interface{}{"itemIconGroupId": public.ID})); len(list) != 1 {
		t.Fatal(list)
	}
	if list := responseList(t, request(t, 1, false, items.GetListByGroupId, map[string]interface{}{"itemIconGroupId": private.ID})); len(list) != 1 {
		t.Fatal(list)
	}
}

func TestAllPrivateVisitorPageDoesNotMutateDatabase(t *testing.T) {
	db := testDB(t)
	private := createGroup(t, db, 1, 1)
	card := models.ItemIcon{UserId: 1, ItemIconGroupId: int(private.ID), Title: "private"}
	db.Create(&card)
	if list := responseList(t, request(t, 1, true, (&ItemIconGroup{}).GetList, nil)); len(list) != 0 {
		t.Fatal(list)
	}
	var count int64
	db.Model(&models.ItemIconGroup{}).Count(&count)
	if count != 1 {
		t.Fatalf("Public read created a group: %d", count)
	}
	var persisted models.ItemIcon
	db.First(&persisted, card.ID)
	if persisted.ItemIconGroupId != int(private.ID) {
		t.Fatal("Private card was moved")
	}
}

func TestGroupEditPreservesSettingsAndRejectsOtherOwners(t *testing.T) {
	db := testDB(t)
	group := createGroup(t, db, 1, 1)
	res := request(t, 2, false, (&ItemIconGroup{}).Edit, map[string]interface{}{"id": group.ID, "title": "stolen"})
	if res["code"] != float64(1005) {
		t.Fatal(res)
	}
	res = request(t, 1, false, (&ItemIconGroup{}).Edit, map[string]interface{}{"id": group.ID, "title": "renamed"})
	if res["code"] != float64(0) {
		t.Fatal(res)
	}
	var persisted models.ItemIconGroup
	db.First(&persisted, group.ID)
	if persisted.Title != "renamed" || persisted.PublicVisitHide != 1 || persisted.CardStyle != group.CardStyle {
		t.Fatal(persisted)
	}
}

func TestCardCannotBeStolenOrMovedIntoAnotherUsersGroup(t *testing.T) {
	db := testDB(t)
	own := createGroup(t, db, 1, 0)
	other := createGroup(t, db, 2, 0)
	card := models.ItemIcon{UserId: 2, ItemIconGroupId: int(other.ID), Title: "other"}
	db.Create(&card)
	res := request(t, 1, false, (&ItemIcon{}).Edit, map[string]interface{}{"id": card.ID, "title": "stolen", "itemIconGroupId": own.ID})
	if res["code"] != float64(1005) {
		t.Fatal(res)
	}
	res = request(t, 1, false, (&ItemIcon{}).Edit, map[string]interface{}{"title": "new", "itemIconGroupId": other.ID})
	if res["code"] != float64(1005) {
		t.Fatal(res)
	}
	var persisted models.ItemIcon
	db.First(&persisted, card.ID)
	if persisted.UserId != 2 || persisted.Title != "other" {
		t.Fatal(persisted)
	}
}

func TestBulkInsertRejectsForeignGroupWithoutPartialWrites(t *testing.T) {
	db := testDB(t)
	own := createGroup(t, db, 1, 0)
	other := createGroup(t, db, 2, 0)
	res := request(t, 1, false, (&ItemIcon{}).AddMultiple, []map[string]interface{}{{"title": "own", "itemIconGroupId": own.ID}, {"title": "other", "itemIconGroupId": other.ID}})
	if res["code"] != float64(1005) {
		t.Fatal(res)
	}
	var count int64
	db.Model(&models.ItemIcon{}).Count(&count)
	if count != 0 {
		t.Fatal("Bulk operation partially inserted cards")
	}
}

func TestGroupSettingsCanBeSavedAndCleared(t *testing.T) {
	db := testDB(t)
	group := createGroup(t, db, 1, 1)
	res := request(t, 1, false, (&ItemIconGroup{}).Edit, map[string]interface{}{"id": group.ID, "title": "public", "publicVisitHide": 0, "cardStyle": "", "sort": 0})
	if res["code"] != float64(0) {
		t.Fatal(res)
	}
	var persisted models.ItemIconGroup
	db.First(&persisted, group.ID)
	if persisted.PublicVisitHide != 0 || persisted.CardStyle != "" || persisted.Sort != 0 {
		t.Fatal(persisted)
	}
}

func TestGroupRejectsInvalidStyleAndVisibility(t *testing.T) {
	db := testDB(t)
	group := createGroup(t, db, 1, 1)
	for _, fields := range []map[string]interface{}{{"cardStyle": "oops"}, {"cardStyle": `{"style":999}`}, {"publicVisitHide": 2}} {
		fields["id"] = group.ID
		fields["title"] = "invalid"
		if res := request(t, 1, false, (&ItemIconGroup{}).Edit, fields); res["code"] == float64(0) {
			t.Fatal(res)
		}
	}
	var persisted models.ItemIconGroup
	db.First(&persisted, group.ID)
	if persisted.PublicVisitHide != 1 || persisted.Title != "group" {
		t.Fatal(persisted)
	}
}
