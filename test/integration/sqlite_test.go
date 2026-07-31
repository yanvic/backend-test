package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autoparts/backend-test/internal/domain"
	"github.com/autoparts/backend-test/internal/repository/sqlite"
)

func setupRepo(t *testing.T) *sqlite.Repo {
	t.Helper()
	repo, err := sqlite.New(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() {
		// SQLite :memory: is automatically released when connection closes
	})
	return repo
}

func newTestPart(id, name string) *domain.Part {
	now := time.Now()
	return &domain.Part{
		ID:                id,
		Name:              name,
		Category:          "test",
		CurrentStock:      100,
		MinimumStock:      50,
		AverageDailySales: 10,
		LeadTimeDays:      5,
		UnitCost:          25.0,
		CriticalityLevel:  3,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func TestSQLiteCreateAndGetByID(t *testing.T) {
	repo := setupRepo(t)

	part := newTestPart("p1", "Part One")
	err := repo.Create(t.Context(), part)
	require.NoError(t, err)

	got, err := repo.GetByID(t.Context(), "p1")
	require.NoError(t, err)

	assert.Equal(t, "p1", got.ID)
	assert.Equal(t, "Part One", got.Name)
	assert.Equal(t, "test", got.Category)
	assert.Equal(t, domain.Stock(100), got.CurrentStock)
	assert.Equal(t, domain.Stock(50), got.MinimumStock)
	assert.Equal(t, domain.DailySales(10), got.AverageDailySales)
	assert.Equal(t, domain.LeadTimeDays(5), got.LeadTimeDays)
	assert.Equal(t, 25.0, got.UnitCost)
	assert.Equal(t, domain.CriticalityLevel(3), got.CriticalityLevel)
	assert.False(t, got.CreatedAt.IsZero())
	assert.False(t, got.UpdatedAt.IsZero())
}

func TestSQLiteCreateDuplicate(t *testing.T) {
	repo := setupRepo(t)

	part := newTestPart("dup", "Duplicate")
	err := repo.Create(t.Context(), part)
	require.NoError(t, err)

	err = repo.Create(t.Context(), part)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "já existe")
}

func TestSQLiteGetByIDNotFound(t *testing.T) {
	repo := setupRepo(t)

	_, err := repo.GetByID(t.Context(), "nonexistent")
	assert.Error(t, err)
}

func TestSQLiteGetAllEmpty(t *testing.T) {
	repo := setupRepo(t)

	parts, err := repo.GetAll(t.Context())
	require.NoError(t, err)
	assert.Empty(t, parts)
}

func TestSQLiteGetAllMultiple(t *testing.T) {
	repo := setupRepo(t)

	for i, name := range []string{"A", "B", "C"} {
		id := "p" + string(rune('1'+i))
		err := repo.Create(t.Context(), newTestPart(id, name))
		require.NoError(t, err)
	}

	parts, err := repo.GetAll(t.Context())
	require.NoError(t, err)
	assert.Len(t, parts, 3)
}

func TestSQLiteUpdate(t *testing.T) {
	repo := setupRepo(t)

	part := newTestPart("upd", "Before")
	err := repo.Create(t.Context(), part)
	require.NoError(t, err)

	part.Name = "After"
	part.CurrentStock = 200
	part.UnitCost = 30.0
	part.UpdatedAt = time.Now()

	err = repo.Update(t.Context(), part)
	require.NoError(t, err)

	got, err := repo.GetByID(t.Context(), "upd")
	require.NoError(t, err)
	assert.Equal(t, "After", got.Name)
	assert.Equal(t, domain.Stock(200), got.CurrentStock)
	assert.Equal(t, 30.0, got.UnitCost)
}

func TestSQLiteUpdateNotFound(t *testing.T) {
	repo := setupRepo(t)

	part := newTestPart("ghost", "Ghost")
	err := repo.Update(t.Context(), part)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "não encontrada")
}

func TestSQLiteDelete(t *testing.T) {
	repo := setupRepo(t)

	err := repo.Create(t.Context(), newTestPart("del", "Delete Me"))
	require.NoError(t, err)

	err = repo.Delete(t.Context(), "del")
	require.NoError(t, err)

	_, err = repo.GetByID(t.Context(), "del")
	assert.Error(t, err)
}

func TestSQLiteDeleteNotFound(t *testing.T) {
	repo := setupRepo(t)

	err := repo.Delete(t.Context(), "ghost")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "não encontrada")
}

func TestSQLitePartWithUnitCost(t *testing.T) {
	repo := setupRepo(t)

	part := newTestPart("cost", "With Cost")
	part.UnitCost = 49.99
	err := repo.Create(t.Context(), part)
	require.NoError(t, err)

	got, err := repo.GetByID(t.Context(), "cost")
	require.NoError(t, err)
	assert.Equal(t, 49.99, got.UnitCost)
}
