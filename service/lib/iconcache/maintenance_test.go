package iconcache

import (
	"path/filepath"
	"sun-panel/models"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestReferencesSkipMissingOptionalTables(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		t.Fatal(err)
	}
	// Mirrors a fresh install, where the notice table does not exist.
	if err := db.AutoMigrate(&models.User{}, &models.SystemSetting{}, &models.ItemIcon{}, &models.UserConfig{}, &models.ModuleConfig{}); err != nil {
		t.Fatal(err)
	}
	db.Create(&models.ItemIcon{Title: "card", IconJson: `{"itemType":2,"src":"https://icons.example/a.png"}`})
	sources, _, err := references(db)
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range sources {
		if source == "https://icons.example/a.png" {
			return
		}
	}
	t.Fatal("card icon reference missing", sources)
}
