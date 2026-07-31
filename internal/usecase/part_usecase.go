package usecase

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/autoparts/backend-test/internal/domain"
	"github.com/autoparts/backend-test/internal/repository"
)

type PartUseCase struct {
	repo repository.PartRepository
}

func NewPartUseCase(repo repository.PartRepository) *PartUseCase {
	return &PartUseCase{repo: repo}
}

type CreatePartInput struct {
	ID                string
	Name              string
	Category          string
	CurrentStock      int
	MinimumStock      int
	AverageDailySales float64
	LeadTimeDays      int
	UnitCost          float64
	CriticalityLevel  int
}

func (uc *PartUseCase) Create(ctx context.Context, input CreatePartInput) (*domain.Part, error) {
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = uuid.New().String()
	}

	if strings.TrimSpace(input.Name) == "" {
		return nil, fmt.Errorf(domain.MsgNameRequired)
	}

	criticality := domain.CriticalityLevel(input.CriticalityLevel)
	if !criticality.IsValid() {
		return nil, fmt.Errorf(domain.MsgCriticalityOutOfRange, input.CriticalityLevel)
	}

	leadTime := domain.LeadTimeDays(input.LeadTimeDays)
	if !leadTime.IsValid() {
		return nil, fmt.Errorf(domain.MsgLeadTimeNegative, input.LeadTimeDays)
	}

	dailySales := domain.DailySales(input.AverageDailySales)
	if !dailySales.IsValid() {
		return nil, fmt.Errorf(domain.MsgAvgDailySalesNegative, input.AverageDailySales)
	}

	now := time.Now()
	part := &domain.Part{
		ID:                id,
		Name:              strings.TrimSpace(input.Name),
		Category:          input.Category,
		CurrentStock:      domain.Stock(input.CurrentStock),
		MinimumStock:      domain.Stock(input.MinimumStock),
		AverageDailySales: dailySales,
		LeadTimeDays:      leadTime,
		UnitCost:          input.UnitCost,
		CriticalityLevel:  criticality,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := uc.repo.Create(ctx, part); err != nil {
		return nil, err
	}
	return part, nil
}

func (uc *PartUseCase) GetByID(ctx context.Context, id string) (*domain.Part, error) {
	return uc.repo.GetByID(ctx, id)
}

type UpdatePartInput struct {
	Name              *string
	Category          *string
	CurrentStock      *int
	MinimumStock      *int
	AverageDailySales *float64
	LeadTimeDays      *int
	UnitCost          *float64
	CriticalityLevel  *int
}

func (uc *PartUseCase) Update(ctx context.Context, id string, input UpdatePartInput) (*domain.Part, error) {
	part, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		part.Name = *input.Name
	}
	if input.Category != nil {
		part.Category = *input.Category
	}
	if input.CurrentStock != nil {
		part.CurrentStock = domain.Stock(*input.CurrentStock)
	}
	if input.MinimumStock != nil {
		part.MinimumStock = domain.Stock(*input.MinimumStock)
	}
	if input.AverageDailySales != nil {
		sales := domain.DailySales(*input.AverageDailySales)
		if !sales.IsValid() {
			return nil, fmt.Errorf(domain.MsgAvgDailySalesInvalid)
		}
		part.AverageDailySales = sales
	}
	if input.LeadTimeDays != nil {
		lt := domain.LeadTimeDays(*input.LeadTimeDays)
		if !lt.IsValid() {
			return nil, fmt.Errorf(domain.MsgLeadTimeInvalid)
		}
		part.LeadTimeDays = lt
	}
	if input.UnitCost != nil {
		part.UnitCost = *input.UnitCost
	}
	if input.CriticalityLevel != nil {
		crit := domain.CriticalityLevel(*input.CriticalityLevel)
		if !crit.IsValid() {
			return nil, fmt.Errorf(domain.MsgCriticalityInvalid)
		}
		part.CriticalityLevel = crit
	}

	part.UpdatedAt = time.Now()

	if err := uc.repo.Update(ctx, part); err != nil {
		return nil, err
	}
	return part, nil
}

func (uc *PartUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

type ListPartsFilter struct {
	Category     string
	NeedsRestock *bool
	Page         int
	PageSize     int
}

type ListPartsResult struct {
	Parts      []*domain.Part
	TotalCount int
}

func (uc *PartUseCase) List(ctx context.Context, filter ListPartsFilter) (*ListPartsResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 50
	}

	parts, err := uc.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	filtered := make([]*domain.Part, 0)
	for _, p := range parts {
		if filter.Category != "" && p.Category != filter.Category {
			continue
		}
		if filter.NeedsRestock != nil && p.NeedsRestock() != *filter.NeedsRestock {
			continue
		}
		filtered = append(filtered, p)
	}

	totalCount := len(filtered)

	start := (filter.Page - 1) * filter.PageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + filter.PageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	return &ListPartsResult{
		Parts:      filtered[start:end],
		TotalCount: totalCount,
	}, nil
}

type PriorityItem struct {
	Part    *domain.Part
	Urgency domain.UrgencyScore
}

func (uc *PartUseCase) GetRestockPriorities(ctx context.Context) ([]PriorityItem, error) {
	parts, err := uc.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]PriorityItem, len(parts))
	for i, p := range parts {
		items[i] = PriorityItem{Part: p, Urgency: p.Urgency()}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Urgency != items[j].Urgency {
			return items[i].Urgency > items[j].Urgency
		}
		if items[i].Part.CriticalityLevel != items[j].Part.CriticalityLevel {
			return items[i].Part.CriticalityLevel > items[j].Part.CriticalityLevel
		}
		if items[i].Part.AverageDailySales != items[j].Part.AverageDailySales {
			return items[i].Part.AverageDailySales > items[j].Part.AverageDailySales
		}
		return items[i].Part.Name < items[j].Part.Name
	})

	return items, nil
}
