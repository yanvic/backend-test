package http

type CreatePartRequest struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Category          string  `json:"category"`
	CurrentStock      int     `json:"currentStock"`
	MinimumStock      int     `json:"minimumStock"`
	AverageDailySales float64 `json:"averageDailySales"`
	LeadTimeDays      int     `json:"leadTimeDays"`
	UnitCost          float64 `json:"unitCost"`
	CriticalityLevel  int     `json:"criticalityLevel"`
}

type UpdatePartRequest struct {
	Name              *string  `json:"name,omitempty"`
	Category          *string  `json:"category,omitempty"`
	CurrentStock      *int     `json:"currentStock,omitempty"`
	MinimumStock      *int     `json:"minimumStock,omitempty"`
	AverageDailySales *float64 `json:"averageDailySales,omitempty"`
	LeadTimeDays      *int     `json:"leadTimeDays,omitempty"`
	UnitCost          *float64 `json:"unitCost,omitempty"`
	CriticalityLevel  *int     `json:"criticalityLevel,omitempty"`
}

type PartResponse struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Category          string  `json:"category"`
	CurrentStock      int     `json:"currentStock"`
	MinimumStock      int     `json:"minimumStock"`
	AverageDailySales float64 `json:"averageDailySales"`
	LeadTimeDays      int     `json:"leadTimeDays"`
	UnitCost          float64 `json:"unitCost"`
	CriticalityLevel  int     `json:"criticalityLevel"`
	ProjectedStock    int     `json:"projectedStock"`
	NeedsRestock      bool    `json:"needsRestock"`
	UrgencyScore      float64 `json:"urgencyScore"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

type PriorityItemResponse struct {
	PartID           string  `json:"partId"`
	Name             string  `json:"name"`
	Category         string  `json:"category"`
	CurrentStock     int     `json:"currentStock"`
	ProjectedStock   int     `json:"projectedStock"`
	MinimumStock     int     `json:"minimumStock"`
	AverageDailySales float64 `json:"averageDailySales"`
	LeadTimeDays     int     `json:"leadTimeDays"`
	UnitCost         float64 `json:"unitCost"`
	CriticalityLevel int     `json:"criticalityLevel"`
	NeedsRestock     bool    `json:"needsRestock"`
	UrgencyScore     float64 `json:"urgencyScore"`
}

type PrioritiesResponse struct {
	Priorities []PriorityItemResponse `json:"priorities"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
