package unit

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/autoparts/backend-test/internal/repository/memory"
	"github.com/autoparts/backend-test/internal/usecase"
)

func TestConcurrentCreateAndGetPriorities(t *testing.T) {
	repo := memory.New()
	uc := usecase.NewPartUseCase(repo)
	ctx := context.Background()

	const numGoroutines = 50
	const partsPerGoroutine = 20

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*partsPerGoroutine)

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < partsPerGoroutine; i++ {
				_, err := uc.Create(ctx, usecase.CreatePartInput{
					ID:                fmt.Sprintf("conc-%d-%d", gid, i),
					Name:              fmt.Sprintf("Concurrent %d-%d", gid, i),
					Category:          "concurrent",
					CurrentStock:      100,
					MinimumStock:      50,
					AverageDailySales: 5,
					LeadTimeDays:      3,
					UnitCost:          10.0,
					CriticalityLevel:  3,
				})
				if err != nil {
					errors <- err
				}
			}
		}(g)
	}

	for g := 0; g < numGoroutines/2; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			priorities, err := uc.GetRestockPriorities(ctx)
			if err != nil {
				errors <- err
				return
			}
			for _, p := range priorities {
				_ = p.Urgency
				_ = p.Part.Name
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent error: %v", err)
	}

	all, err := uc.GetRestockPriorities(ctx)
	assert.NoError(t, err)
	assert.Len(t, all, numGoroutines*partsPerGoroutine,
		"should have all created parts after concurrent operations")
}
