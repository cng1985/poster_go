package tests

import (
	"context"
	"testing"

	"gorm.io/gorm"

	"poster.app/studio/internal/modules/account"
	"poster.app/studio/internal/modules/studio/asset"
	"poster.app/studio/internal/modules/studio/poster"
	"poster.app/studio/internal/modules/studio/shared"
	"poster.app/studio/internal/platform/config"
	"poster.app/studio/internal/platform/domain"
	apperrors "poster.app/studio/internal/platform/errors"
	"poster.app/studio/internal/platform/storage"
	"poster.app/studio/internal/testkit"
)

func TestDeleteReferencedAsset(t *testing.T) {
	db, cfg := testkit.DB(t)
	ctx := registerOwner(t, db, cfg, "Erin")
	files, err := storage.NewLocal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	repo := asset.NewRepository(db)
	resource := asset.NewResource(repo, files, cfg)
	ownerID, err := domain.ActorID(ctx)
	if err != nil {
		t.Fatal(err)
	}
	created, err := repo.Create(ctx, &asset.Asset{
		OwnerID: ownerID, StorageKey: "assets/0100/leaf.png", OriginalName: "leaf.png",
		ContentType: "image/png", ByteSize: 8, Width: 96, Height: 96,
	})
	if err != nil {
		t.Fatal(err)
	}
	posters := poster.NewResource(poster.NewRepository(db), resource)
	doc := shared.Document{
		Width: 1080, Height: 1440, Background: "#112233",
		Layers: []shared.Layer{
			{
				ID: "t1", Type: "text", Name: "title", X: 80, Y: 80, W: 640, H: 140, Opacity: 1,
				Text: "还在", FontSize: 64, FontWeight: 700, FontFamily: "serif", Color: "#ffffff", Align: "left",
			},
			{
				ID: "img", Type: "image", Name: "leaf", X: 80, Y: 260, W: 400, H: 280, Opacity: 1,
				AssetID: created.Data.ID,
			},
		},
	}
	saved, err := posters.Create(ctx, &poster.PosterRequest{Title: "带图", Document: doc})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resource.Delete(ctx, &asset.AssetRequest{ID: created.Data.ID}); err == nil {
		t.Fatal("expected referenced asset to stay")
	} else if appErr, ok := apperrors.As(err); !ok || appErr.Code != 10005 {
		t.Fatalf("delete error = %v", err)
	}
	if _, err := resource.One(ctx, &asset.AssetRequest{ID: created.Data.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := posters.Delete(ctx, &poster.PosterRequest{ID: saved.Data.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := resource.Delete(ctx, &asset.AssetRequest{ID: created.Data.ID}); err != nil {
		t.Fatal(err)
	}
}

func registerOwner(t *testing.T, db *gorm.DB, cfg *config.Config, username string) context.Context {
	t.Helper()
	repo, err := account.NewRepository(db, cfg)
	if err != nil {
		t.Fatal(err)
	}
	created, err := account.NewResource(repo, cfg).Create(domain.WithClientIP(context.Background(), "10.1.0.9"), &account.AccountRegisterRequest{
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
