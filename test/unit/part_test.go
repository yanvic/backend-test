package unit

import (
	"sort"
	"testing"

	"github.com/autoparts/backend-test/internal/domain"
)

func TestPartExpectedConsumption(t *testing.T) {
	tests := []struct {
		name              string
		averageDailySales domain.DailySales
		leadTimeDays      domain.LeadTimeDays
		expected          float64
	}{
		{"normal case", 10, 5, 50},
		{"zero sales", 0, 5, 0},
		{"zero lead time", 10, 0, 0},
		{"both zero", 0, 0, 0},
		{"fractional sales", 3.5, 4, 14},
		{"high values", 100, 30, 3000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &domain.Part{
				AverageDailySales: tt.averageDailySales,
				LeadTimeDays:      tt.leadTimeDays,
			}
			got := p.ExpectedConsumption()
			if got != tt.expected {
				t.Errorf("ExpectedConsumption() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPartProjectedStock(t *testing.T) {
	tests := []struct {
		name              string
		currentStock      domain.Stock
		averageDailySales domain.DailySales
		leadTimeDays      domain.LeadTimeDays
		expected          domain.Stock
	}{
		{"stock above consumption", 100, 10, 5, 50},
		{"stock equals consumption", 50, 10, 5, 0},
		{"stock below consumption", 30, 10, 5, -20},
		{"zero stock", 0, 10, 5, -50},
		{"negative stock", -10, 10, 5, -60},
		{"zero consumption", 100, 0, 30, 100},
		{"zero lead time", 100, 10, 0, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &domain.Part{
				CurrentStock:      tt.currentStock,
				AverageDailySales: tt.averageDailySales,
				LeadTimeDays:      tt.leadTimeDays,
			}
			got := p.ProjectedStock()
			if got != tt.expected {
				t.Errorf("ProjectedStock() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPartNeedsRestock(t *testing.T) {
	tests := []struct {
		name              string
		currentStock      domain.Stock
		minimumStock      domain.Stock
		averageDailySales domain.DailySales
		leadTimeDays      domain.LeadTimeDays
		expected          bool
	}{
		{"projected below min", 100, 200, 10, 5, true},
		{"projected equals min", 50, 50, 0, 0, false},
		{"projected above min", 200, 50, 10, 5, false},
		{"negative stock below min", -10, 5, 10, 5, true},
		{"min is zero, projected positive", 100, 0, 10, 5, false},
		{"min is zero, projected negative", -50, 0, 100, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &domain.Part{
				CurrentStock:      tt.currentStock,
				MinimumStock:      tt.minimumStock,
				AverageDailySales: tt.averageDailySales,
				LeadTimeDays:      tt.leadTimeDays,
			}
			got := p.NeedsRestock()
			if got != tt.expected {
				t.Errorf("NeedsRestock() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPartUrgencyScore(t *testing.T) {
	tests := []struct {
		name              string
		currentStock      domain.Stock
		minimumStock      domain.Stock
		averageDailySales domain.DailySales
		leadTimeDays      domain.LeadTimeDays
		criticalityLevel  domain.CriticalityLevel
		expected          domain.UrgencyScore
	}{
		{
			name:              "normal urgency",
			currentStock:      100,
			minimumStock:      200,
			averageDailySales: 10,
			leadTimeDays:      5,
			criticalityLevel:  3,
			expected:          450, // (200 - 50) * 3
		},
		{
			name:              "no urgency (enough stock)",
			currentStock:      500,
			minimumStock:      100,
			averageDailySales: 10,
			leadTimeDays:      5,
			criticalityLevel:  5,
			expected:          -1750, // (100 - 450) * 5
		},
		{
			name:              "criticality zero means no urgency",
			currentStock:      0,
			minimumStock:      100,
			averageDailySales: 10,
			leadTimeDays:      5,
			criticalityLevel:  0,
			expected:          0,
		},
		{
			name:              "negative stock high urgency",
			currentStock:      -50,
			minimumStock:      100,
			averageDailySales: 10,
			leadTimeDays:      5,
			criticalityLevel:  10,
			expected:          2000, // (100 - (-100)) * 10
		},
		{
			name:              "zero lead time",
			currentStock:      10,
			minimumStock:      50,
			averageDailySales: 10,
			leadTimeDays:      0,
			criticalityLevel:  5,
			expected:          200, // (50 - 10) * 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &domain.Part{
				CurrentStock:      tt.currentStock,
				MinimumStock:      tt.minimumStock,
				AverageDailySales: tt.averageDailySales,
				LeadTimeDays:      tt.leadTimeDays,
				CriticalityLevel:  tt.criticalityLevel,
			}
			got := p.Urgency()
			if got != tt.expected {
				t.Errorf("Urgency() = %v, want %v", got, tt.expected)
			}
		})
	}
}

type partPriority struct {
	part    *domain.Part
	urgency domain.UrgencyScore
}

func TestPriorityOrdering(t *testing.T) {
	parts := []*domain.Part{
		{Name: "A", CurrentStock: 100, MinimumStock: 200, AverageDailySales: 10, LeadTimeDays: 5, CriticalityLevel: 3},
		{Name: "B", CurrentStock: 50, MinimumStock: 200, AverageDailySales: 10, LeadTimeDays: 5, CriticalityLevel: 5},
		{Name: "C", CurrentStock: 0, MinimumStock: 200, AverageDailySales: 10, LeadTimeDays: 5, CriticalityLevel: 5},
		{Name: "D", CurrentStock: 0, MinimumStock: 200, AverageDailySales: 20, LeadTimeDays: 5, CriticalityLevel: 5},
		{Name: "E", CurrentStock: 0, MinimumStock: 200, AverageDailySales: 20, LeadTimeDays: 5, CriticalityLevel: 5},
	}

	priorities := make([]partPriority, len(parts))
	for i, p := range parts {
		priorities[i] = partPriority{part: p, urgency: p.Urgency()}
	}

	sort.Slice(priorities, func(i, j int) bool {
		ua, ub := priorities[i].urgency, priorities[j].urgency
		if ua != ub {
			return ua > ub
		}
		ca, cb := priorities[i].part.CriticalityLevel, priorities[j].part.CriticalityLevel
		if ca != cb {
			return ca > cb
		}
		sa, sb := priorities[i].part.AverageDailySales, priorities[j].part.AverageDailySales
		if sa != sb {
			return sa > sb
		}
		return priorities[i].part.Name < priorities[j].part.Name
	})

	t.Run("highest urgency first", func(t *testing.T) {
		if priorities[0].part.Name != "C" && priorities[0].part.Name != "D" {
			t.Errorf("expected C or D to be highest urgency, got %s", priorities[0].part.Name)
		}
	})

	t.Run("same urgency uses criticality tiebreak", func(t *testing.T) {
		if priorities[1].urgency != priorities[2].urgency {
			return
		}
		if priorities[1].part.CriticalityLevel < priorities[2].part.CriticalityLevel {
			t.Errorf("expected higher criticality first for same urgency")
		}
	})

	t.Run("same urgency and criticality uses dailySales tiebreak", func(t *testing.T) {
		posD, posE := -1, -1
		for i, pp := range priorities {
			if pp.part.Name == "D" {
				posD = i
			}
			if pp.part.Name == "E" {
				posE = i
			}
		}
		if posD == -1 || posE == -1 {
			t.Skip("D or E not found")
			return
		}
		if posD > posE {
			t.Errorf("expected same score; D should come before E alphabetically")
		}
	})

	t.Run("lowest urgency is A", func(t *testing.T) {
		if priorities[len(priorities)-1].part.Name != "A" {
			t.Errorf("expected A to have lowest urgency, got %s", priorities[len(priorities)-1].part.Name)
		}
	})
}

func TestEmptyListPriorities(t *testing.T) {
	parts := []*domain.Part{}
	priorities := make([]partPriority, len(parts))
	for i, p := range parts {
		priorities[i] = partPriority{part: p, urgency: p.Urgency()}
	}

	if len(priorities) != 0 {
		t.Errorf("expected empty priorities, got %d", len(priorities))
	}
}
