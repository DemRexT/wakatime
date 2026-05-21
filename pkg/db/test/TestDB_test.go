package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"wakatime/pkg/db"

	"github.com/go-pg/pg/v10"
)

func TestCommonRepoUserInsertGetDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skip db test in short mode")
	}

	ctx := context.Background()
	dbo, _ := Setup(t)
	repo := db.NewCommonRepo(dbo.DB)

	user, err := repo.AddUser(ctx, &db.User{
		Username:      "ivan",
		WakatimeLogin: "ivanlogin9087",
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
	dbo, _ := Setup(t)
	repo := db.NewCommonRepo(dbo.DB)

	user, err := repo.AddUser(ctx, &db.User{
		Username:      "ivan1",
		WakatimeLogin: "ivanlogin1",
		WakatimeToken: []byte("waka_13231"),
		StatusID:      db.StatusEnabled,
	})
	if err != nil {
		t.Fatalf("add user failed: %v", err)
	}

	t.Cleanup(func() {
		_, _ = repo.DeleteUser(ctx, user.ID)
	})

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

	t.Cleanup(func() {
		_, _ = repo.DeleteStat(ctx, stat.ID)
	})

	duplicate := *stat
	duplicate.ID = 0
	duplicate.TotalSeconds = 2000

	_, err = repo.AddStat(ctx, &duplicate)
	if err == nil {
		t.Fatal("expected duplicate stat error, got nil")
	}

	var pgErr pg.Error
	if !errors.As(err, &pgErr) || !pgErr.IntegrityViolation() {
		t.Fatalf("expected integrity violation, got %v", err)
	}
}

func TestUpdateUserDoesNotMutateID(t *testing.T) {
	if testing.Short() {
		t.Skip("skip db test in short mode")
	}
	ctx := context.Background()
	dbo, _ := Setup(t)
	repo := db.NewCommonRepo(dbo.DB)

	user, err := repo.AddUser(ctx, &db.User{
		Username:      "ivan",
		WakatimeLogin: "ivanlogin789",
		WakatimeToken: []byte("waka_132312"),
		StatusID:      db.StatusEnabled,
	})
	if err != nil {
		t.Fatalf("add user failed: %v", err)
	}

	originalID := user.ID

	t.Cleanup(func() {
		_, cleanupErr := repo.DeleteUser(ctx, user.ID)
		if cleanupErr != nil {
			t.Logf("cleanup user failed: %v", err)
		}
	})

	user.Username = "Ivan12"

	ok, err := repo.UpdateUser(ctx, user)
	if err != nil {
		t.Fatalf("update user failed %v", err)
	}
	if !ok {
		t.Fatalf("expected user to be updated")
	}

	updateUser, err := repo.UserByID(ctx, originalID)
	if err != nil {
		t.Fatalf("get user failed: %v", err)
	}

	if updateUser.ID != originalID {
		t.Fatalf("expected user id %d, got %d", originalID, updateUser.ID)
	}

	if user.Username != updateUser.Username {
		t.Fatalf("expected username %q, got %q", user.Username, updateUser.Username)
	}
}

func TestWithEnabledOnlyHidesDisabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skip db test in short mode")
	}
	ctx := context.Background()
	dbo, _ := Setup(t)
	repo := db.NewCommonRepo(dbo.DB)

	user, err := repo.AddUser(ctx, &db.User{
		Username:      "ivan",
		WakatimeLogin: "ivanlogin578",
		WakatimeToken: []byte("waka_132312"),
		StatusID:      db.StatusDisabled,
	})
	if err != nil {
		t.Fatalf("add user failed: %v", err)
	}

	t.Cleanup(func() {
		_, cleanupErr := repo.DeleteUser(ctx, user.ID)
		if cleanupErr != nil {
			t.Logf("cleanup user failed: %v", err)
		}
	})

	users, err := repo.UsersByFilters(ctx, nil, db.PagerDefault)
	if err != nil {
		t.Fatalf("get users failed: %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}

	if users[0].Username != "ivan" {
		t.Fatalf("expected username %q, got %q", "ivan", users[0].Username)
	}

	users, err = repo.WithEnabledOnly().UsersByFilters(ctx, nil, db.PagerDefault)
	if err != nil {
		t.Fatalf("get users failed: %v", err)
	}

	if len(users) != 0 {
		t.Fatalf("expected 0 user, got %d", len(users))
	}
}

func TestUpdateStatDoesNotMutateID(t *testing.T) {
	if testing.Short() {
		t.Skip("skip db test in short mode")
	}
	ctx := context.Background()
	dbo, _ := Setup(t)
	repo := db.NewCommonRepo(dbo.DB)
	user, err := repo.AddUser(ctx, &db.User{
		Username:      "ivan",
		WakatimeLogin: "ivanlogin7969",
		WakatimeToken: []byte("waka_132312"),
		StatusID:      db.StatusEnabled,
	})
	if err != nil {
		t.Fatalf("add user failed: %v", err)
	}

	t.Cleanup(func() {
		_, cleanupErr := repo.DeleteUser(ctx, user.ID)
		if cleanupErr != nil {
			t.Logf("cleanup user failed: %v", err)
		}
	})

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
		t.Fatalf("add stat failed: %v", err)
	}

	originalID := stat.ID

	t.Cleanup(func() {
		_, cleanupErr := repo.DeleteStat(ctx, stat.ID)
		if cleanupErr != nil {
			t.Logf("cleanup stat failed: %v", err)
		}
	})

	stat.TotalSeconds = 2000

	ok, err := repo.UpdateStat(ctx, stat)
	if err != nil {
		t.Fatalf("update stat failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected stat to be updated")
	}

	updatedStat, err := repo.StatByID(ctx, originalID)
	if err != nil {
		t.Fatalf("get stat failed: %v", err)
	}

	if updatedStat.ID != originalID {
		t.Fatalf("expected stat ID %d, got %d", originalID, updatedStat.ID)
	}

	if updatedStat.TotalSeconds != 2000 {
		t.Fatalf("expected total seconds %d, got %d", 2000, updatedStat.TotalSeconds)
	}
}
