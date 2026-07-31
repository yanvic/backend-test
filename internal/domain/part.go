package domain

import "time"

type Part struct {
	ID                string
	Name              string
	Category          string
	CurrentStock      Stock
	MinimumStock      Stock
	AverageDailySales DailySales
	LeadTimeDays      LeadTimeDays
	UnitCost          float64
	CriticalityLevel  CriticalityLevel
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (l CriticalityLevel) IsValid() bool {
	return l >= 1 && l <= 5
}

func (l LeadTimeDays) IsValid() bool {
	return l >= 0
}

func (s DailySales) IsValid() bool {
	return s >= 0
}

func (p *Part) ExpectedConsumption() float64 {
	return float64(p.AverageDailySales) * float64(p.LeadTimeDays)
}

func (p *Part) ProjectedStock() Stock {
	expected := int(p.ExpectedConsumption())
	return p.CurrentStock.Project(expected)
}

func (p *Part) NeedsRestock() bool {
	return p.ProjectedStock().NeedsRestock(p.MinimumStock)
}

func (p *Part) Urgency() UrgencyScore {
	return p.ProjectedStock().UrgencyScore(p.MinimumStock, p.CriticalityLevel)
}
