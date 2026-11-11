package poster

import (
	"context"
	"testing"

	"gorm.io/gorm"

	"poster.app/studio/internal/modules/account"
	"poster.app/studio/internal/modules/studio/asset"
	"poster.app/studio/internal/modules/studio/shared"
	"poster.app/studio/internal/platform/config"
	"poster.app/studio/internal/platform/domain"
	"poster.app/studio/internal/platform/storage"
	"poster.app/studio/internal/testkit"
)

func TestPosterVersionOwnerAndZeroOpacity(t *testing.T) {
	db, cfg := testkit.DB(t)
	owner := register(t, db, cfg, "Cara")
	other := register(t, db, cfg, "Dora")
	files, err := storage.NewLocal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	assets := asset.NewResource(asset.NewRepository(db), files, cfg)
	resource := NewResource(NewRepository(db), assets)
	doc := sampleDoc()
	doc.Layers[0].Opacity = 0
	created, err := resource.Create(owner, &PosterRequest{Title: "清晨", Document: doc})
	if err != nil {
		t.Fatal(err)
	}
	if created.Data.Version != 1 || created.Data.Document.Layers[0].Opacity != 0 {
		t.Fatalf("version=%d opacity=%v", created.Data.Version, created.Data.Document.Layers[0].Opacity)
	}
	updated, err := resource.Update(owner, &PosterRequest{
		ID: created.Data.ID, Title: "清晨", Version: 1, Document: created.Data.Document,
	})
	if err != nil || updated.Data.Version != 2 {
		t.Fatal(err, updated.Data.Version)
	}
	if _, err := resource.Update(owner, &PosterRequest{
		ID: created.Data.ID, Title: "旧版本", Version: 1, Document: created.Data.Document,
	}); err == nil {
		t.Fatal("expected version conflict")
	}
	page, err := resource.Search(other, &PosterSearch{})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 0 || page.Data == nil || len(page.Data) != 0 {
		t.Fatalf("%+v", page)
	}
	if _, err := resource.One(other, &PosterRequest{ID: created.Data.ID}); err == nil {
		t.Fatal("other owner can read poster")
	}
	if _, err := resource.One(owner, &PosterRequest{ID: 999}); err == nil {
		t.Fatal("expected missing poster")
	}
	if _, err := resource.Delete(owner, &PosterRequest{ID: created.Data.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := resource.Delete(owner, &PosterRequest{ID: created.Data.ID}); err == nil {
		t.Fatal("repeat delete should be not found")
	}
	if _, err := resource.Search(owner, &PosterSearch{Page: -1}); err == nil {
		t.Fatal("expected invalid page")
	}
	bad := db.Create(&Poster{OwnerID: 999999, Title: "x", Width: 1080, Height: 1440, CoverColor: "#111111", Document: `{"width":1080}`, Version: 1})
	if bad.Error == nil {
		t.Fatal("foreign key did not reject unknown owner")
	}
}

func sampleDoc() shared.Document {
	return shared.Document{
		Width: 1080, Height: 1440, Background: "#112233",
		Layers: []shared.Layer{{
			ID: "t1", Type: "text", Name: "title", X: 80, Y: 120, W: 800, H: 180, Opacity: 1,
			Text: "你好", FontSize: 72, FontWeight: 700, FontFamily: "serif", Color: "#ffffff", Align: "left",
		}},
	}
}

func register(t *testing.T, db *gorm.DB, cfg *config.Config, username string) context.Context {
	t.Helper()
	repo, err := account.NewRepository(db, cfg)
	if err != nil {
		t.Fatal(err)
	}
	created, err := account.NewResource(repo, cfg).Create(domain.WithClientIP(context.Background(), "10.0.0.8"), &account.AccountRegisterRequest{
		Username: username, Password: "secret123", DisplayName: username,
	})
	if err != nil {
		t.Fatal(err)
	}
	scope, err := account.NewSessionResolver(db).Resolve(context.Background(), created.Data.Token)
	if err != nil {
		t.Fatal(err)
	}
	return domain.WithScope(context.Background(), scope)
}
