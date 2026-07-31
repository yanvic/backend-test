package unit

import (
	"context"
	"fmt"
	"testing"

	"github.com/autoparts/backend-test/internal/domain"
	"github.com/autoparts/backend-test/internal/repository/memory"
	"github.com/autoparts/backend-test/internal/usecase"
)

func BenchmarkGetRestockPriorities(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			repo := memory.New()
			uc := usecase.NewPartUseCase(repo)
			ctx := context.Background()

			for i := 0; i < size; i++ {
				_, err := uc.Create(ctx, usecase.CreatePartInput{
					ID:                fmt.Sprintf("p-%d", i),
					Name:              fmt.Sprintf("Part %d", i),
					Category:          "test",
					CurrentStock:      i * 10,
					MinimumStock:      100,
					AverageDailySales: float64(i%20 + 1),
					LeadTimeDays:      i%10 + 1,
					UnitCost:          float64(i) * 1.5,
					CriticalityLevel:  i%5 + 1,
				})
				if err != nil {
					b.Fatal(err)
				}
			}

			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				_, err := uc.GetRestockPriorities(ctx)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkCreatePart(b *testing.B) {
	repo := memory.New()
	uc := usecase.NewPartUseCase(repo)
	ctx := context.Background()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := uc.Create(ctx, usecase.CreatePartInput{
			ID:                fmt.Sprintf("bench-%d", n),
			Name:              fmt.Sprintf("Bench Part %d", n),
			Category:          "bench",
			CurrentStock:      100,
			MinimumStock:      50,
			AverageDailySales: 5,
			LeadTimeDays:      3,
			UnitCost:          10.0,
			CriticalityLevel:  3,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUrgencyScore(b *testing.B) {
	p := &domain.Part{
		CurrentStock:      100,
		MinimumStock:      200,
		AverageDailySales: 10,
		LeadTimeDays:      5,
		CriticalityLevel:  3,
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = p.Urgency()
	}
}
