package service

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"xcoding/apps/project/internal/models"
	projectv1 "xcoding/gen/go/project/v1"
)

// setupTestDB creates an in-memory sqlite DB and migrates models.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Project{}, &models.ProjectMember{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

// seedProjects seeds projects and membership for a given user.
func seedProjects(t *testing.T, db *gorm.DB, ownerID uint64, memberID uint64, total int) {
	t.Helper()
	// create projects with alternating owners, and membership entries
	for i := 0; i < total; i++ {
		p := models.Project{Name: "p" + string(rune('a'+i)), OwnerID: ownerID}
		if err := db.Create(&p).Error; err != nil {
			t.Fatalf("create project: %v", err)
		}
		// Add member to half of projects
		if i%2 == 0 {
			m := models.ProjectMember{ProjectID: p.ID, UserID: memberID, Role: int32(projectv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_MEMBER)}
			if err := db.Create(&m).Error; err != nil {
				t.Fatalf("create member: %v", err)
			}
		}
	}
}

func TestListProjects_CountMatchesFilter(t *testing.T) {
	db := setupTestDB(t)
	svc := NewProjectService(db)

	// Seed 20 projects owned by ownerID=1; user 2 is member of 10 projects (even indices)
	seedProjects(t, db, 1, 2, 20)

	ctx := context.Background()

	// Case 1: owner view (ownerID=1), all=false → should see 20 projects; pageSize=7 → pages = ceil(20/7)=3
	ownerID := uint64(1)
	projects, total, totalPages, err := svc.ListProjects(ctx, 1, 7, &ownerID, false)
	if err != nil {
		t.Fatalf("ListProjects owner: %v", err)
	}
	if total != 20 {
		t.Fatalf("expected total=20, got %d", total)
	}
	if totalPages != 3 {
		t.Fatalf("expected totalPages=3, got %d", totalPages)
	}
	if len(projects) != 7 {
		t.Fatalf("expected first page size=7, got %d", len(projects))
	}

	// Case 2: member view (ownerID=2), all=false → should see owned (0) + joined (10) projects
	memberID := uint64(2)
	_, total, totalPages, err = svc.ListProjects(ctx, 1, 5, &memberID, false)
	if err != nil {
		t.Fatalf("ListProjects member: %v", err)
	}
	if total != 10 {
		t.Fatalf("expected total=10, got %d", total)
	}
	if totalPages != 2 {
		t.Fatalf("expected totalPages=2, got %d", totalPages)
	}

	// Case 3: unauthenticated (ownerID=nil), all=false → returns all (20) as per current logic
	_, total, totalPages, err = svc.ListProjects(ctx, 1, 10, nil, false)
	if err != nil {
		t.Fatalf("ListProjects nil owner: %v", err)
	}
	if total != 20 {
		t.Fatalf("expected total=20, got %d", total)
	}
	if totalPages != 2 {
		t.Fatalf("expected totalPages=2, got %d", totalPages)
	}
}
