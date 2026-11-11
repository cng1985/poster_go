package account

import (
	"context"
	"testing"

	"poster.app/studio/internal/platform/domain"
	apperrors "poster.app/studio/internal/platform/errors"
	"poster.app/studio/internal/testkit"
)

func TestRegisterLoginAndRollback(t *testing.T) {
	db, cfg := testkit.DB(t)
	repo, err := NewRepository(db, cfg)
	if err != nil {
		t.Fatal(err)
	}
	resource := NewResource(repo, cfg)
	ctx := domain.WithClientIP(context.Background(), "127.0.0.1")
	created, err := resource.Create(ctx, &AccountRegisterRequest{
		Username: "Ada", Password: "secret123", DisplayName: "艾达",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Data.Token == "" || created.Data.Account.ID == 0 {
		t.Fatalf("%+v", created.Data)
	}
	_, err = resource.Create(ctx, &AccountRegisterRequest{Username: "Ada", Password: "secret123"})
	if err == nil {
		t.Fatal("expected duplicate username")
	}
	if appErr, ok := apperrors.As(err); !ok || appErr.Code != 10001 {
		t.Fatal(err)
	}
	var accounts int64
	if err := db.Model(&Account{}).Count(&accounts).Error; err != nil {
		t.Fatal(err)
	}
	if accounts != 1 {
		t.Fatalf("accounts = %d", accounts)
	}
	bad, err := resource.Login(ctx, &AccountLoginRequest{Username: "Ada", Password: "wrong123"})
	if err == nil || bad.Data.Token != "" {
		t.Fatal("expected login failure")
	}
	okLogin, err := resource.Login(ctx, &AccountLoginRequest{Username: "Ada", Password: "secret123"})
	if err != nil {
		t.Fatal(err)
	}
	scope, err := NewSessionResolver(db).Resolve(ctx, okLogin.Data.Token)
	if err != nil {
		t.Fatal(err)
	}
	authed := domain.WithScope(ctx, scope)
	me, err := resource.One(authed, &AccountRequest{})
	if err != nil || me.Data.DisplayName != "艾达" {
		t.Fatal(err, me.Data)
	}
	if _, err := resource.Logout(authed, &AccountRequest{}); err != nil {
		t.Fatal(err)
	}
	if _, err := NewSessionResolver(db).Resolve(ctx, okLogin.Data.Token); err == nil {
		t.Fatal("revoked token still resolved")
	}

	if err := db.Exec(`create trigger abort_session before insert on session begin select raise(abort, 'fail session'); end;`).Error; err != nil {
		t.Fatal(err)
	}
	before := accounts
	_, err = resource.Create(ctx, &AccountRegisterRequest{Username: "Bea", Password: "secret123"})
	if err == nil {
		t.Fatal("expected session failure")
	}
	if err := db.Model(&Account{}).Count(&accounts).Error; err != nil {
		t.Fatal(err)
	}
	if accounts != before {
		t.Fatalf("register leaked account: %d -> %d", before, accounts)
	}
}
