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

func TestCommonRepoUserInsertGetDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skip db test in short mode")
	}

	ctx := context.Background()
	dbo := newTestDB(t)
	repo := db.NewCommonRepo(dbo)

	user, err := repo.AddUser(ctx, &db.User{
		Username:      "ivan",
		WakatimeLogin: "ivanlogin",
		WakatimeToken: []byte("test_token"),
		StatusID:      db.StatusEnabled,
	})
	if err != nil {
		t.Fatalf("add user failed: %v", err)
	}

	t.Cleanup(func() {
		_, _ = dbo.ModelContext(ctx, &db.User{ID: user.ID}).WherePK().Delete()
	})

	got, err := repo.UserByID(ctx, user.ID, repo.FullUser())
	if err != nil {
		t.Fatalf("get user failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected user, got nil")
	}
	if got.Username != user.Username {
		t.Fatalf("expected username %q, got %q", user.Username, got.Username)
	}

	deleted, err := repo.DeleteUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("delete user failed: %v", err)
	}
	if !deleted {
		t.Fatal("expected user to be deleted")
	}

	got, err = repo.UserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("get deleted user failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected deleted user to still exist with deleted status")
	}
	if got.StatusID != db.StatusDeleted {
		t.Fatalf("expected status %d, got %d", db.StatusDeleted, got.StatusID)
	}
}
