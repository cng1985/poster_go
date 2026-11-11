package generation

import (
	"context"
	"strings"
	"testing"

	"poster.app/studio/internal/modules/studio/shared"
	"poster.app/studio/internal/platform/ai"
	"poster.app/studio/internal/platform/domain"
	"poster.app/studio/internal/platform/limit"
	"poster.app/studio/internal/testkit"
	"poster.app/studio/migrations"
)

func TestLocalGeneration(t *testing.T) {
	db, _ := testkit.DB(t)
	if err := migrations.Apply(db); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`insert into account(username, password_hash, display_name, status, created_at, updated_at) values ('gen', 'x', 'gen', 'active', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`).Error; err != nil {
		t.Fatal(err)
	}
	var owner int64
	if err := db.Raw("select id from account where username = 'gen'").Scan(&owner).Error; err != nil {
		t.Fatal(err)
	}
	resource := &Resource{
		repo:   NewRepository(db),
		models: localModel{},
		limit:  limit.NewWindow(10, timeMinute),
	}
	ctx := domain.WithScope(context.Background(), domain.Scope{ActorID: owner})
	result, err := resource.Create(ctx, &GenerationRequest{
		Prompt: "春日咖啡，暖色，手写感", Style: "editorial", SizePreset: "story",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Data.Provider != "local" || result.Data.Document.Width != 1080 || len(result.Data.Document.Layers) == 0 {
		t.Fatalf("%+v", result.Data)
	}
	if !strings.Contains(result.Data.Document.Layers[2].Text, "春日咖啡") && !strings.Contains(result.Data.Message, "本地") {
		t.Fatalf("document=%+v message=%s", result.Data.Document, result.Data.Message)
	}
}

type localModel struct{}

func (localModel) Enabled() bool       { return false }
func (localModel) FallbackLocal() bool { return true }
func (localModel) Provider() string    { return "openai" }
func (localModel) Model() string       { return "none" }
func (localModel) Complete(context.Context, []ai.Message) (string, error) {
	return "", nil
}

var _ = shared.Document{}
