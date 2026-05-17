package test

import (
	"context"
	"testing"
	"time"

	"wakatime/pkg/db"
)

func TestCommonRepoUserInsertGetDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skip db test in short mode")
	}

	ctx := context.Background()
	dbo, err := setup()
	if err != nil {
		t.Fatalf("failed connect: %v", err)
	}
	repo := db.NewCommonRepo(dbo)

	user, err := repo.AddUser(ctx, &db.User{
		Username:      "ivan",
		WakatimeLogin: "ivanlogin",
		WakatimeToken: []byte("waka_132312"),
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

	users, err := repo.WithEnabledOnly().UsersByFilters(ctx, nil, db.PagerDefault)
	if err != nil {
		t.Fatalf("get enabled users failed: %v", err)
	}

	for _, item := range users {
		if item.ID == user.ID {
			t.Fatal("deleted user should not be returned by enabled-only filter")
		}
	}
}

func TestCommonRepoStatUniquePeriod(t *testing.T) {
	if testing.Short() {
		t.Skip("skip db test in short mode")
	}
	ctx := context.Background()
	dbo, err := setup()
	if err != nil {
		t.Fatalf("failed connect: %v", err)
	}
	repo := db.NewCommonRepo(dbo)

	user, err := repo.AddUser(ctx, &db.User{
		Username:      "ivan1",
		WakatimeLogin: "ivanlogin1",
		WakatimeToken: []byte("waka_13231"),
		StatusID:      db.StatusEnabled,
	})
	if err != nil {
		t.Fatalf("add user failed: %v", err)
	}

	stat, err := repo.AddStat(ctx, &db.Stat{
		UserID:              user.ID,
		Period:              "week",
		PeriodStart:         time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:           time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC),
		TotalSeconds:        1000,
		DailyAverageSeconds: 143,
		FetchedAt:           time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	duplicate := *stat
	duplicate.ID = 3
	duplicate.TotalSeconds = 2000

	_, err = repo.AddStat(ctx, &duplicate)
	if err == nil {
		t.Fatal("expected duplicate stat error, got nil")
	}

	_, err = repo.DeleteStat(ctx, stat.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.DeleteUser(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
}
