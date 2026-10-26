package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/compucueva/compucueva/internal/api"
	"github.com/compucueva/compucueva/internal/db"
)

type testEnv struct {
	server  *api.Server
	queries *db.Queries
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	conn, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	schema, err := os.ReadFile("../sql/schema.sql")
	if err != nil {
		t.Fatalf("failed to read schema.sql: %v", err)
	}

	if _, err := conn.ExecContext(context.Background(), string(schema)); err != nil {
		t.Fatalf("failed to apply schema.sql: %v", err)
	}

	queries := db.New(conn)
	return &testEnv{
		server:  api.NewServer(queries),
		queries: queries,
	}
}

func TestCreateTask(t *testing.T) {
	env := setupTestEnv(t)

	t.Run("valid payload", func(t *testing.T) {
		payload := `{"title":"Diagnose motherboard","description":"No POST beep"}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d. body: %s", rec.Code, rec.Body.String())
		}

		var task db.Task
		if err := json.NewDecoder(rec.Body).Decode(&task); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if task.Title != "Diagnose motherboard" || task.Status != "TODO" || task.Priority != "NORMAL" {
			t.Errorf("unexpected task defaults: %+v", task)
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(`{"title":`))
		rec := httptest.NewRecorder()

		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
	})
}

func TestListTasks(t *testing.T) {
	env := setupTestEnv(t)

	t.Run("empty list returns json array not null", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
		rec := httptest.NewRecorder()

		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		if strings.TrimSpace(rec.Body.String()) != "[]" {
			t.Fatalf("expected empty JSON array [], got %s", rec.Body.String())
		}
	})

	t.Run("populated list", func(t *testing.T) {
		_, err := env.queries.CreateTask(context.Background(), db.CreateTaskParams{
			Title:       "Task 1",
			Description: "Description 1",
		})
		if err != nil {
			t.Fatalf("failed to seed task: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
		rec := httptest.NewRecorder()

		env.server.Routes().ServeHTTP(rec, req)

		var tasks []db.Task
		if err := json.NewDecoder(rec.Body).Decode(&tasks); err != nil {
			t.Fatalf("failed to decode tasks: %v", err)
		}
		if len(tasks) != 1 {
			t.Fatalf("expected 1 task, got %d", len(tasks))
		}
	})
}

func TestGetTaskByID(t *testing.T) {
	env := setupTestEnv(t)

	seeded, err := env.queries.CreateTask(context.Background(), db.CreateTaskParams{
		Title:       "PSU Test",
		Description: "Check 12V rail",
	})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	t.Run("found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/tasks/%d", seeded.ID), nil)
		rec := httptest.NewRecorder()

		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks/99999", nil)
		rec := httptest.NewRecorder()

		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("invalid non-numeric id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks/abc", nil)
		rec := httptest.NewRecorder()

		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
	})
}

func TestUpdateTaskStatus(t *testing.T) {
	env := setupTestEnv(t)

	seeded, err := env.queries.CreateTask(context.Background(), db.CreateTaskParams{
		Title:       "RAM Testing",
		Description: "Run memtest86",
	})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	t.Run("valid transition", func(t *testing.T) {
		payload := `{"status":"DOING"}`
		url := fmt.Sprintf("/api/tasks/%d/status", seeded.ID)
		req := httptest.NewRequest(http.MethodPatch, url, bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var updated db.Task
		if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if updated.Status != "DOING" {
			t.Errorf("expected status %q, got %q", "DOING", updated.Status)
		}
	})

	t.Run("non-existent task returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/tasks/99999/status", bytes.NewBufferString(`{"status":"DONE"}`))
		rec := httptest.NewRecorder()

		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("invalid status rejected by check constraint", func(t *testing.T) {
		payload := `{"status":"INVALID_STATUS"}`
		url := fmt.Sprintf("/api/tasks/%d/status", seeded.ID)
		req := httptest.NewRequest(http.MethodPatch, url, bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
	})
}

func TestUpdateTask(t *testing.T) {
	env := setupTestEnv(t)

	seeded, err := env.queries.CreateTask(context.Background(), db.CreateTaskParams{
		Title:       "Original Title",
		Description: "Original Description",
	})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	t.Run("valid update", func(t *testing.T) {
		payload := `{"title":"Updated Title","description":"Updated Description","priority":"HIGH"}`
		url := fmt.Sprintf("/api/tasks/%d", seeded.ID)
		req := httptest.NewRequest(http.MethodPut, url, bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var updated db.Task
		if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if updated.Title != "Updated Title" || updated.Priority != "HIGH" {
			t.Errorf("task fields did not update properly: %+v", updated)
		}
	})

	t.Run("invalid priority rejected by check constraint", func(t *testing.T) {
		payload := `{"title":"T","description":"D","priority":"URGENT"}`
		url := fmt.Sprintf("/api/tasks/%d", seeded.ID)
		req := httptest.NewRequest(http.MethodPut, url, bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()

		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500 on invalid priority, got %d", rec.Code)
		}
	})

	t.Run("non-existent task returns 404", func(t *testing.T) {
		payload := `{"title":"T","description":"D","priority":"LOW"}`
		req := httptest.NewRequest(http.MethodPut, "/api/tasks/99999", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()

		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
	})
}

func TestDeleteTask(t *testing.T) {
	env := setupTestEnv(t)

	seeded, err := env.queries.CreateTask(context.Background(), db.CreateTaskParams{
		Title:       "Delete Me",
		Description: "To be removed",
	})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	url := fmt.Sprintf("/api/tasks/%d", seeded.ID)

	// First delete succeeds
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	rec := httptest.NewRecorder()
	env.server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}

	// Second delete on same ID must return 404
	req2 := httptest.NewRequest(http.MethodDelete, url, nil)
	rec2 := httptest.NewRecorder()
	env.server.Routes().ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 on second delete, got %d", rec2.Code)
	}
}
