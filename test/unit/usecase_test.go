package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autoparts/backend-test/internal/repository/memory"
	"github.com/autoparts/backend-test/internal/usecase"
)

func setupUseCase() *usecase.PartUseCase {
	return usecase.NewPartUseCase(memory.New())
}

func validInput(id string) usecase.CreatePartInput {
	return usecase.CreatePartInput{
		ID:                id,
		Name:              "Test Part",
		Category:          "engine",
		CurrentStock:      100,
		MinimumStock:      50,
		AverageDailySales: 10,
		LeadTimeDays:      5,
		UnitCost:          25.0,
		CriticalityLevel:  3,
	}
}

func TestUseCaseCreateValid(t *testing.T) {
	uc := setupUseCase()

	part, err := uc.Create(t.Context(), validInput("p1"))
	require.NoError(t, err)

	assert.Equal(t, "p1", part.ID)
	assert.Equal(t, "Test Part", part.Name)
	assert.Equal(t, "engine", part.Category)
	assert.False(t, part.CreatedAt.IsZero())
	assert.False(t, part.UpdatedAt.IsZero())
}

func TestUseCaseCreateAutoUUID(t *testing.T) {
	uc := setupUseCase()

	input := validInput("")
	part, err := uc.Create(t.Context(), input)
	require.NoError(t, err)

	assert.NotEmpty(t, part.ID)
	assert.Len(t, part.ID, 36) // UUID v4 format: 8-4-4-4-12
	assert.Contains(t, part.ID, "-")
}

func TestUseCaseCreateWithoutName(t *testing.T) {
	uc := setupUseCase()

	input := validInput("p1")
	input.Name = ""

	_, err := uc.Create(t.Context(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nome é obrigatório")
}

func TestUseCaseCreateCriticalityTooLow(t *testing.T) {
	uc := setupUseCase()

	input := validInput("p1")
	input.CriticalityLevel = 0

	_, err := uc.Create(t.Context(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deve estar entre 1 e 5")
	assert.Contains(t, err.Error(), "recebeu 0")
}

func TestUseCaseCreateCriticalityTooHigh(t *testing.T) {
	uc := setupUseCase()

	input := validInput("p1")
	input.CriticalityLevel = 6

	_, err := uc.Create(t.Context(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deve estar entre 1 e 5")
}

func TestUseCaseCreateLeadTimeNegative(t *testing.T) {
	uc := setupUseCase()

	input := validInput("p1")
	input.LeadTimeDays = -1

	_, err := uc.Create(t.Context(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "leadTimeDays deve ser >= 0")
}

func TestUseCaseCreateSalesNegative(t *testing.T) {
	uc := setupUseCase()

	input := validInput("p1")
	input.AverageDailySales = -1

	_, err := uc.Create(t.Context(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "averageDailySales deve ser >= 0")
}

func TestUseCaseCreateDuplicate(t *testing.T) {
	uc := setupUseCase()

	_, err := uc.Create(t.Context(), validInput("dup"))
	require.NoError(t, err)

	_, err = uc.Create(t.Context(), validInput("dup"))
	assert.Error(t, err)
}

func TestUseCaseGetByID(t *testing.T) {
	uc := setupUseCase()

	_, err := uc.Create(t.Context(), validInput("g1"))
	require.NoError(t, err)

	part, err := uc.GetByID(t.Context(), "g1")
	require.NoError(t, err)
	assert.Equal(t, "Test Part", part.Name)
}

func TestUseCaseGetByIDNotFound(t *testing.T) {
	uc := setupUseCase()

	_, err := uc.GetByID(t.Context(), "missing")
	assert.Error(t, err)
}

func TestUseCaseUpdatePartial(t *testing.T) {
	uc := setupUseCase()

	_, err := uc.Create(t.Context(), validInput("u1"))
	require.NoError(t, err)

	newName := "Updated"
	newStock := 200
	part, err := uc.Update(t.Context(), "u1", usecase.UpdatePartInput{
		Name:         &newName,
		CurrentStock: &newStock,
	})
	require.NoError(t, err)

	assert.Equal(t, "Updated", part.Name)
	assert.Equal(t, 200, int(part.CurrentStock))
	assert.Equal(t, "engine", part.Category) // unchanged
}

func TestUseCaseUpdateValidation(t *testing.T) {
	uc := setupUseCase()

	_, err := uc.Create(t.Context(), validInput("u2"))
	require.NoError(t, err)

	badLeadTime := -5
	_, err = uc.Update(t.Context(), "u2", usecase.UpdatePartInput{
		LeadTimeDays: &badLeadTime,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "leadTimeDays deve ser >= 0")
}

func TestUseCaseUpdateNotFound(t *testing.T) {
	uc := setupUseCase()

	newName := "Ghost"
	_, err := uc.Update(t.Context(), "ghost", usecase.UpdatePartInput{
		Name: &newName,
	})
	assert.Error(t, err)
}

func TestUseCaseDelete(t *testing.T) {
	uc := setupUseCase()

	_, err := uc.Create(t.Context(), validInput("d1"))
	require.NoError(t, err)

	err = uc.Delete(t.Context(), "d1")
	require.NoError(t, err)

	_, err = uc.GetByID(t.Context(), "d1")
	assert.Error(t, err)
}

func TestUseCaseDeleteNotFound(t *testing.T) {
	uc := setupUseCase()

	err := uc.Delete(t.Context(), "ghost")
	assert.Error(t, err)
}

func TestUseCaseListWithFilter(t *testing.T) {
	uc := setupUseCase()

	create := func(id, category string, stock int) {
		input := validInput(id)
		input.Category = category
		input.CurrentStock = stock
		_, err := uc.Create(t.Context(), input)
		require.NoError(t, err)
	}

	create("l1", "cat-a", 10)  // needs restock (stock=10, min=50)
	create("l2", "cat-b", 200) // no restock
	create("l3", "cat-a", 5)   // needs restock

	result, err := uc.List(t.Context(), usecase.ListPartsFilter{
		Category: "cat-a",
	})
	require.NoError(t, err)
	assert.Len(t, result.Parts, 2)
	assert.Equal(t, 2, result.TotalCount)

	result, err = uc.List(t.Context(), usecase.ListPartsFilter{
		NeedsRestock: boolPtr(true),
	})
	require.NoError(t, err)
	assert.Len(t, result.Parts, 2) // l1 and l3
}

func TestUseCaseListPagination(t *testing.T) {
	uc := setupUseCase()

	for i := 0; i < 5; i++ {
		id := "pg" + string(rune('0'+i))
		_, err := uc.Create(t.Context(), validInput(id))
		require.NoError(t, err)
	}

	result, err := uc.List(t.Context(), usecase.ListPartsFilter{
		Page:     1,
		PageSize: 2,
	})
	require.NoError(t, err)
	assert.Len(t, result.Parts, 2)
	assert.Equal(t, 5, result.TotalCount)

	result, err = uc.List(t.Context(), usecase.ListPartsFilter{
		Page:     99,
		PageSize: 2,
	})
	require.NoError(t, err)
	assert.Empty(t, result.Parts)
	assert.Equal(t, 5, result.TotalCount)
}

func TestUseCaseGetRestockPriorities(t *testing.T) {
	uc := setupUseCase()

	create := func(id string, stock int, sales float64, crit int) {
		input := validInput(id)
		input.CurrentStock = stock
		input.AverageDailySales = sales
		input.CriticalityLevel = crit
		_, err := uc.Create(t.Context(), input)
		require.NoError(t, err)
	}

	create("pri1", 10, 10, 1)  // low urgency
	create("pri2", 0, 20, 5)    // high urgency

	priorities, err := uc.GetRestockPriorities(t.Context())
	require.NoError(t, err)
	assert.Len(t, priorities, 2)
	assert.Equal(t, "pri2", priorities[0].Part.ID) // highest urgency first
}

func TestUseCaseGetRestockPrioritiesEmpty(t *testing.T) {
	uc := setupUseCase()

	priorities, err := uc.GetRestockPriorities(t.Context())
	require.NoError(t, err)
	assert.Empty(t, priorities)
}

func boolPtr(b bool) *bool { return &b }
