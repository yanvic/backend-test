package unit

import (
	"testing"

	"github.com/autoparts/backend-test/internal/domain"
)

func FuzzUrgencyScore(f *testing.F) {
	f.Add(100, 200, 10, 5, 3)
	f.Add(100, 50, 10, 5, 5)
	f.Add(0, 100, 10, 5, 1)
	f.Add(-10, 5, 100, 3, 2)
	f.Add(50, 50, 0, 0, 0)

	f.Fuzz(func(t *testing.T, currentStock, minimumStock, averageDailySalesInt, leadTimeDaysInt, criticalityLevel int) {
		if leadTimeDaysInt < 0 {
			leadTimeDaysInt = 0
		}
		if criticalityLevel < 0 {
			criticalityLevel = 0
		}

		p := &domain.Part{
			CurrentStock:      domain.Stock(currentStock),
			MinimumStock:      domain.Stock(minimumStock),
			AverageDailySales: domain.DailySales(averageDailySalesInt),
			LeadTimeDays:      domain.LeadTimeDays(leadTimeDaysInt),
			CriticalityLevel:  domain.CriticalityLevel(criticalityLevel),
		}

		urgency := p.Urgency()

		projected := p.ProjectedStock()
		if projected >= p.MinimumStock && urgency > 0 {
			t.Errorf("urgency should be <= 0 when projected stock is sufficient: urgency=%v, projected=%v, min=%v",
				urgency, projected, p.MinimumStock)
		}

		if p.CriticalityLevel == 0 && urgency != 0 {
			t.Errorf("urgency should be 0 when criticality is 0: got %v", urgency)
		}

		urgency2 := p.Urgency()
		if urgency != urgency2 {
			t.Errorf("Urgency() is not deterministic: %v vs %v", urgency, urgency2)
		}
	})
}
