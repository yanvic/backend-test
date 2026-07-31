package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/autoparts/backend-test/internal/domain"
	"github.com/autoparts/backend-test/internal/usecase"
)

type Handler struct {
	partUseCase *usecase.PartUseCase
}

func NewHandler(partUseCase *usecase.PartUseCase) *Handler {
	return &Handler{partUseCase: partUseCase}
}

func (h *Handler) CreatePart(w http.ResponseWriter, r *http.Request) {
	var req CreatePartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, domain.MsgInvalidBody)
		return
	}

	input := usecase.CreatePartInput{
		ID:                req.ID,
		Name:              req.Name,
		Category:          req.Category,
		CurrentStock:      req.CurrentStock,
		MinimumStock:      req.MinimumStock,
		AverageDailySales: req.AverageDailySales,
		LeadTimeDays:      req.LeadTimeDays,
		UnitCost:          req.UnitCost,
		CriticalityLevel:  req.CriticalityLevel,
	}

	part, err := h.partUseCase.Create(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toPartResponse(part))
}

func (h *Handler) GetPart(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, domain.MsgIDRequired)
		return
	}

	part, err := h.partUseCase.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toPartResponse(part))
}

func (h *Handler) ListParts(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	needsRestockStr := r.URL.Query().Get("needsRestock")

	var needsRestock *bool
	if needsRestockStr == "true" {
		v := true
		needsRestock = &v
	} else if needsRestockStr == "false" {
		v := false
		needsRestock = &v
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))

	result, err := h.partUseCase.List(r.Context(), usecase.ListPartsFilter{
		Category:     category,
		NeedsRestock: needsRestock,
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]PartResponse, len(result.Parts))
	for i, p := range result.Parts {
		response[i] = toPartResponse(p)
	}

	if response == nil {
		response = []PartResponse{}
	}

	w.Header().Set("X-Total-Count", strconv.Itoa(result.TotalCount))
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) UpdatePart(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, domain.MsgIDRequired)
		return
	}

	var req UpdatePartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, domain.MsgInvalidBody)
		return
	}

	input := usecase.UpdatePartInput{
		Name:              req.Name,
		Category:          req.Category,
		CurrentStock:      req.CurrentStock,
		MinimumStock:      req.MinimumStock,
		AverageDailySales: req.AverageDailySales,
		LeadTimeDays:      req.LeadTimeDays,
		UnitCost:          req.UnitCost,
		CriticalityLevel:  req.CriticalityLevel,
	}

	part, err := h.partUseCase.Update(r.Context(), id, input)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toPartResponse(part))
}

func (h *Handler) DeletePart(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, domain.MsgIDRequired)
		return
	}

	if err := h.partUseCase.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetRestockPriorities(w http.ResponseWriter, r *http.Request) {
	priorities, err := h.partUseCase.GetRestockPriorities(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	items := make([]PriorityItemResponse, len(priorities))
	for i, item := range priorities {
		items[i] = PriorityItemResponse{
			PartID:           item.Part.ID,
			Name:             item.Part.Name,
			Category:         item.Part.Category,
			CurrentStock:     int(item.Part.CurrentStock),
			ProjectedStock:   int(item.Part.ProjectedStock()),
			MinimumStock:     int(item.Part.MinimumStock),
			AverageDailySales: float64(item.Part.AverageDailySales),
			LeadTimeDays:     int(item.Part.LeadTimeDays),
			UnitCost:         item.Part.UnitCost,
			CriticalityLevel: int(item.Part.CriticalityLevel),
			NeedsRestock:     item.Part.NeedsRestock(),
			UrgencyScore:     float64(item.Urgency),
		}
	}

	if items == nil {
		items = []PriorityItemResponse{}
	}

	writeJSON(w, http.StatusOK, PrioritiesResponse{Priorities: items})
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func toPartResponse(p *domain.Part) PartResponse {
	return PartResponse{
		ID:                p.ID,
		Name:              p.Name,
		Category:          p.Category,
		CurrentStock:      int(p.CurrentStock),
		MinimumStock:      int(p.MinimumStock),
		AverageDailySales: float64(p.AverageDailySales),
		LeadTimeDays:      int(p.LeadTimeDays),
		UnitCost:          p.UnitCost,
		CriticalityLevel:  int(p.CriticalityLevel),
		ProjectedStock:    int(p.ProjectedStock()),
		NeedsRestock:      p.NeedsRestock(),
		UrgencyScore:      float64(p.Urgency()),
		CreatedAt:         p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         p.UpdatedAt.Format(time.RFC3339),
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
