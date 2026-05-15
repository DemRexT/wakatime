package test

import (
	"context"
	"testing"
	"wakatime/pkg/db"

	"github.com/go-pg/pg/v10"
)

func newTestDB(t *testing.T) *pg.DB {
	t.Helper()

	db := pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "123",
		Database: "wakatime",
		Addr:     "localhost:5432",
	})

	if err := db.Ping(context.Background()); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func TestUserInsertGetDelete(t *testing.T) {
	ctx := context.Background()
	dbo := newTestDB(t)

	user := &db.User{
		Username:      "ivan",
		WakatimeLogin: "ivanlogin",
		WakatimeToken: []byte("waka_2352525"),
		StatusID:      1,
	}

	_, err := dbo.ModelContext(ctx, user).Insert()
	if err != nil {
		t.Fatalf("user failed: %v", err)
	}

	got := &db.User{ID: user.ID}
	err = dbo.ModelContext(ctx, got).WherePK().Select()
	if err != nil {
		t.Fatalf("get user failed: %v", err)
	}

	if got.Username != user.Username {
		t.Fatalf("expected username %q, got %q", user.Username, got.Username)
	}

	_, err = dbo.ModelContext(ctx, user).WherePK().Delete()
	if err != nil {
		t.Fatalf("delete user failed: %v", err)
	}

	err = dbo.ModelContext(ctx, &db.User{ID: user.ID}).WherePK().Select()
	if err == nil {
		t.Fatalf("expected error after delete, got nil")
	}
}
