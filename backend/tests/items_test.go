package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/compucueva/compucueva/internal/db"
)

func TestItemsCRUD(t *testing.T) {
	env := setupTestEnv(t)

	t.Run("empty list returns json array", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/items", nil)
		rec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "[]" {
			t.Fatalf("expected 200 and [], got %d: %s", rec.Code, rec.Body.String())
		}
	})

	var created db.Item
	t.Run("create and list item", func(t *testing.T) {
		payload := `{"type":"PSU","description":"EVGA 500W 80+"}`
		req := httptest.NewRequest(http.MethodPost, "/api/items", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
			t.Fatalf("failed to decode item: %v", err)
		}
		if created.Type != "PSU" || created.Description == nil || *created.Description != "EVGA 500W 80+" {
			t.Fatalf("unexpected item: %+v", created)
		}

		listReq := httptest.NewRequest(http.MethodGet, "/api/items", nil)
		listRec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(listRec, listReq)

		var items []db.Item
		_ = json.NewDecoder(listRec.Body).Decode(&items)
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
	})

	t.Run("delete item and verify double delete returns 404", func(t *testing.T) {
		url := fmt.Sprintf("/api/items/%d", created.ID)

		delReq := httptest.NewRequest(http.MethodDelete, url, nil)
		delRec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(delRec, delReq)
		if delRec.Code != http.StatusNoContent {
			t.Fatalf("expected 204 on first delete, got %d", delRec.Code)
		}

		delReq2 := httptest.NewRequest(http.MethodDelete, url, nil)
		delRec2 := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(delRec2, delReq2)
		if delRec2.Code != http.StatusNotFound {
			t.Fatalf("expected 404 on second delete, got %d", delRec2.Code)
		}
	})
}

func TestTaskItemAssignmentFlow(t *testing.T) {
	env := setupTestEnv(t)

	desc := "Kingston 240GB"
	item, err := env.queries.CreateItem(t.Context(), db.CreateItemParams{
		Type:        "SSD",
		Description: &desc,
	})
	if err != nil {
		t.Fatalf("failed to seed item: %v", err)
	}

	task, err := env.queries.CreateTask(t.Context(), db.CreateTaskParams{
		Title:       "Upgrade storage",
		Description: "Install SSD",
	})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	t.Run("assign non-existent item fails foreign key", func(t *testing.T) {
		payload := fmt.Sprintf(`{"task_id":%d,"item_id":99999}`, task.ID)
		req := httptest.NewRequest(http.MethodPost, "/api/items/assign", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 on foreign key violation, got %d", rec.Code)
		}
	})

	t.Run("assign, list task items, and unassign", func(t *testing.T) {
		payload := fmt.Sprintf(`{"task_id":%d,"item_id":%d}`, task.ID, item.ID)
		req := httptest.NewRequest(http.MethodPost, "/api/items/assign", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204 on assign, got %d: %s", rec.Code, rec.Body.String())
		}

		// Get task items
		getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/tasks/%d/items", task.ID), nil)
		getRec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(getRec, getReq)
		if getRec.Code != http.StatusOK {
			t.Fatalf("expected 200 on get task items, got %d", getRec.Code)
		}
		var taskItems []db.GetTaskItemsRow
		if err := json.NewDecoder(getRec.Body).Decode(&taskItems); err != nil {
			t.Fatalf("failed to decode task items: %v", err)
		}
		if len(taskItems) != 1 || taskItems[0].ID != item.ID || taskItems[0].Type != "SSD" {
			t.Fatalf("unexpected task items: %+v", taskItems)
		}

		// Unassign item from task
		unassignReq := httptest.NewRequest(http.MethodDelete, "/api/items/unassign", bytes.NewBufferString(payload))
		unassignRec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(unassignRec, unassignReq)
		if unassignRec.Code != http.StatusNoContent {
			t.Fatalf("expected 204 on unassign, got %d", unassignRec.Code)
		}

		// Second unassign returns 404
		unassignReq2 := httptest.NewRequest(http.MethodDelete, "/api/items/unassign", bytes.NewBufferString(payload))
		unassignRec2 := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(unassignRec2, unassignReq2)
		if unassignRec2.Code != http.StatusNotFound {
			t.Fatalf("expected 404 on second unassign, got %d", unassignRec2.Code)
		}
	})
}
