package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httphandler "github.com/autoparts/backend-test/internal/http"
	"github.com/autoparts/backend-test/internal/repository/memory"
	"github.com/autoparts/backend-test/internal/usecase"
)

func setupServer() *httptest.Server {
	repo := memory.New()
	usecase := usecase.NewPartUseCase(repo)
	handler := httphandler.NewHandler(usecase)
	router := httphandler.NewRouter(handler)
	return httptest.NewServer(router)
}

func TestCreatePart(t *testing.T) {
	srv := setupServer()
	defer srv.Close()

	body := map[string]interface{}{
		"id":                "part-1",
		"name":              "Velas de ignição",
		"category":          "motor",
		"currentStock":      100,
		"minimumStock":      200,
		"averageDailySales": 10,
		"leadTimeDays":      5,
		"criticalityLevel":  3,
	}

	b, _ := json.Marshal(body)
	resp, err := http.Post(srv.URL+"/api/v1/parts", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var part map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&part)
	assert.Equal(t, "part-1", part["id"])
	assert.Equal(t, "Velas de ignição", part["name"])
	assert.Equal(t, float64(100), part["currentStock"])
}

func TestCreateDuplicate(t *testing.T) {
	srv := setupServer()
	defer srv.Close()

	body := map[string]interface{}{
		"id":                "part-dup",
		"name":              "Duplicada",
		"currentStock":      10,
		"minimumStock":      20,
		"averageDailySales": 1,
		"leadTimeDays":      2,
		"criticalityLevel":  1,
	}

	b, _ := json.Marshal(body)

	resp1, err := http.Post(srv.URL+"/api/v1/parts", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	resp1.Body.Close()
	assert.Equal(t, http.StatusCreated, resp1.StatusCode)

	resp2, err := http.Post(srv.URL+"/api/v1/parts", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusConflict, resp2.StatusCode)
}

func TestGetPartNotFound(t *testing.T) {
	srv := setupServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/parts/nonexistent")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCRUDFlow(t *testing.T) {
	srv := setupServer()
	defer srv.Close()

	body := map[string]interface{}{
		"id":                "crud-1",
		"name":              "Filtro de óleo",
		"category":          "manutenção",
		"currentStock":      50,
		"minimumStock":      100,
		"averageDailySales": 5,
		"leadTimeDays":      7,
		"criticalityLevel":  4,
	}

	b, _ := json.Marshal(body)
	resp, err := http.Post(srv.URL+"/api/v1/parts", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	resp, err = http.Get(srv.URL + "/api/v1/parts/crud-1")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var part map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&part)
	assert.Equal(t, "Filtro de óleo", part["name"])
	assert.Equal(t, float64(50), part["currentStock"])

	update := map[string]interface{}{
		"currentStock": 80,
	}
	ub, _ := json.Marshal(update)
	req, err := http.NewRequest(http.MethodPut, srv.URL+"/api/v1/parts/crud-1", bytes.NewReader(ub))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err = client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get(srv.URL + "/api/v1/parts/crud-1")
	require.NoError(t, err)
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&part)
	assert.Equal(t, float64(80), part["currentStock"])

	req, err = http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/parts/crud-1", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	resp, err = http.Get(srv.URL + "/api/v1/parts/crud-1")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestRestockPriorities(t *testing.T) {
	srv := setupServer()
	defer srv.Close()

	parts := []map[string]interface{}{
		{"id": "p1", "name": "Pistão", "category": "motor", "currentStock": 100, "minimumStock": 200, "averageDailySales": 10, "leadTimeDays": 5, "criticalityLevel": 5},
		{"id": "p2", "name": "Correia", "category": "motor", "currentStock": 500, "minimumStock": 50, "averageDailySales": 2, "leadTimeDays": 3, "criticalityLevel": 2},
		{"id": "p3", "name": "Óleo", "category": "lubrificantes", "currentStock": 0, "minimumStock": 100, "averageDailySales": 20, "leadTimeDays": 5, "criticalityLevel": 5},
		{"id": "p4", "name": "Filtro de ar", "category": "manutenção", "currentStock": 10, "minimumStock": 150, "averageDailySales": 8, "leadTimeDays": 4, "criticalityLevel": 3},
	}

	for _, p := range parts {
		b, _ := json.Marshal(p)
		resp, err := http.Post(srv.URL+"/api/v1/parts", "application/json", bytes.NewReader(b))
		require.NoError(t, err)
		resp.Body.Close()
	}

	resp, err := http.Get(srv.URL + "/api/v1/restock/priorities")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var wrapper struct {
		Priorities []map[string]interface{} `json:"priorities"`
	}
	json.NewDecoder(resp.Body).Decode(&wrapper)

	priorities := wrapper.Priorities
	require.Len(t, priorities, 4)

	assert.Equal(t, "Óleo", priorities[0]["name"],
		"highest urgency should be first")

	assert.Greater(t, priorities[0]["urgencyScore"], priorities[1]["urgencyScore"],
		"urgency should be descending")

	t.Run("tie-break by criticality", func(t *testing.T) {
		for i := 0; i < len(priorities)-1; i++ {
			if priorities[i]["urgencyScore"] == priorities[i+1]["urgencyScore"] {
				assert.GreaterOrEqual(t, priorities[i]["criticalityLevel"], priorities[i+1]["criticalityLevel"])
			}
		}
	})
}

func TestRestockPrioritiesEmpty(t *testing.T) {
	srv := setupServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/restock/priorities")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var wrapper struct {
		Priorities []interface{} `json:"priorities"`
	}
	json.NewDecoder(resp.Body).Decode(&wrapper)
	assert.Empty(t, wrapper.Priorities)
}

func TestListPartsWithFilters(t *testing.T) {
	srv := setupServer()
	defer srv.Close()

	parts := []map[string]interface{}{
		{"id": "f1", "name": "A", "category": "cat-a", "currentStock": 10, "minimumStock": 100, "averageDailySales": 5, "leadTimeDays": 3, "criticalityLevel": 3},
		{"id": "f2", "name": "B", "category": "cat-b", "currentStock": 200, "minimumStock": 50, "averageDailySales": 2, "leadTimeDays": 1, "criticalityLevel": 1},
	}

	for _, p := range parts {
		b, _ := json.Marshal(p)
		resp, err := http.Post(srv.URL+"/api/v1/parts", "application/json", bytes.NewReader(b))
		require.NoError(t, err)
		resp.Body.Close()
	}

	resp, err := http.Get(srv.URL + "/api/v1/parts?category=cat-a")
	require.NoError(t, err)
	defer resp.Body.Close()

	var filtered []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&filtered)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "A", filtered[0]["name"])

	resp, err = http.Get(srv.URL + "/api/v1/parts?needsRestock=true")
	require.NoError(t, err)
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&filtered)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "A", filtered[0]["name"])
}

func TestHealthCheck(t *testing.T) {
	srv := setupServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestNegativeStock(t *testing.T) {
	srv := setupServer()
	defer srv.Close()

	body := map[string]interface{}{
		"id":                "neg-1",
		"name":              "Negative Stock",
		"currentStock":      -50,
		"minimumStock":      100,
		"averageDailySales": 10,
		"leadTimeDays":      5,
		"criticalityLevel":  5,
	}

	b, _ := json.Marshal(body)
	resp, err := http.Post(srv.URL+"/api/v1/parts", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	resp, err = http.Get(srv.URL + "/api/v1/parts/neg-1")
	require.NoError(t, err)
	defer resp.Body.Close()

	var part map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&part)
	assert.Equal(t, float64(-50), part["currentStock"])
	assert.Equal(t, true, part["needsRestock"])
	assert.Greater(t, part["urgencyScore"], float64(0))
}
