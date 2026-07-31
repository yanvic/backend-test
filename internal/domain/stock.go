package domain

type Stock int
type DailySales float64
type LeadTimeDays int
type CriticalityLevel int
type UrgencyScore float64

func (s Stock) Project(expectedConsumption int) Stock {
	return s - Stock(expectedConsumption)
}

func (projected Stock) NeedsRestock(minimum Stock) bool {
	return projected < minimum
}

func (projected Stock) UrgencyScore(minimum Stock, criticality CriticalityLevel) UrgencyScore {
	deficit := float64(minimum - projected)
	return UrgencyScore(deficit * float64(criticality))
}
