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

func TestCreateAndGetUser(t *testing.T) {
	env := setupTestEnv(t)

	t.Run("create user and fetch by id and username", func(t *testing.T) {
		payload := `{"username":"messi","name":"Lionel Messi","role":"ADMIN"}`
		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var created db.User
		if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
			t.Fatalf("failed to decode user: %v", err)
		}
		if created.Username != "messi" || created.Role == nil || *created.Role != "ADMIN" {
			t.Fatalf("unexpected user fields: %+v", created)
		}

		// Get by ID
		getIDReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d", created.ID), nil)
		getIDRec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(getIDRec, getIDReq)
		if getIDRec.Code != http.StatusOK {
			t.Fatalf("expected status 200 on get by id, got %d", getIDRec.Code)
		}

		// Get by username
		getUserReq := httptest.NewRequest(http.MethodGet, "/api/users/username/messi", nil)
		getUserRec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(getUserRec, getUserReq)
		if getUserRec.Code != http.StatusOK {
			t.Fatalf("expected status 200 on get by username, got %d", getUserRec.Code)
		}

		// Get /api/me via Cloudflare header
		meReq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		meReq.Header.Set("Cf-Access-Authenticated-User-Email", "messi")
		meRec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(meRec, meReq)
		if meRec.Code != http.StatusOK {
			t.Fatalf("expected status 200 on /api/me, got %d", meRec.Code)
		}

		// Get /api/users (ListUsers)
		listReq := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		listRec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(listRec, listReq)
		if listRec.Code != http.StatusOK {
			t.Fatalf("expected status 200 on /api/users, got %d", listRec.Code)
		}
	})

	t.Run("empty username rejected by check constraint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(`{"username":"   ","name":"Test"}`))
		rec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 on empty username, got %d", rec.Code)
		}
	})

	t.Run("duplicate username rejected by unique constraint", func(t *testing.T) {
		payload := `{"username":"messi","name":"Duplicate","role":"USER"}`
		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()

		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected error on duplicate username, got %d", rec.Code)
		}
	})

	t.Run("non-existent user returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/users/99999", nil)
		rec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for unknown id, got %d", rec.Code)
		}

		req2 := httptest.NewRequest(http.MethodGet, "/api/users/username/nobody", nil)
		rec2 := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(rec2, req2)
		if rec2.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for unknown username, got %d", rec2.Code)
		}
	})
}

func TestUserRoleAndTaskAssignmentFlow(t *testing.T) {
	env := setupTestEnv(t)

	roleTech := "TECH"
	u1, err := env.queries.CreateUser(t.Context(), db.CreateUserParams{
		Username: "tech1",
		Name:     "Tech One",
		Role:     &roleTech,
	})
	if err != nil {
		t.Fatalf("failed to seed u1: %v", err)
	}

	u2, err := env.queries.CreateUser(t.Context(), db.CreateUserParams{
		Username: "tech2",
		Name:     "Tech Two",
		Role:     &roleTech,
	})
	if err != nil {
		t.Fatalf("failed to seed u2: %v", err)
	}

	task, err := env.queries.CreateTask(t.Context(), db.CreateTaskParams{
		Title:       "Clean PC",
		Description: "Remove dust",
	})
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	t.Run("filter by role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/users/role/TECH", nil)
		rec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var users []db.User
		if err := json.NewDecoder(rec.Body).Decode(&users); err != nil {
			t.Fatalf("failed to decode users: %v", err)
		}
		if len(users) != 2 {
			t.Fatalf("expected 2 TECH users, got %d", len(users))
		}
	})

	t.Run("assign non-existent user or task fails foreign key", func(t *testing.T) {
		payload := fmt.Sprintf(`{"user_id":99999,"task_id":%d}`, task.ID)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/assign", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected foreign key failure 500, got %d", rec.Code)
		}
	})

	t.Run("assign user to task and verify with-tasks / without-tasks / assignees", func(t *testing.T) {
		payload := fmt.Sprintf(`{"user_id":%d,"task_id":%d}`, u1.ID, task.ID)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/assign", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
		}

		// GetTaskAssignees
		assReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/tasks/%d/assignees", task.ID), nil)
		assRec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(assRec, assReq)

		if assRec.Code != http.StatusOK {
			t.Fatalf("expected 200 on get assignees, got %d: %s", assRec.Code, assRec.Body.String())
		}
		var assignees []db.GetTaskAssigneesRow
		if err := json.NewDecoder(assRec.Body).Decode(&assignees); err != nil {
			t.Fatalf("failed to decode assignees: %v", err)
		}
		if len(assignees) != 1 || assignees[0].Username != "tech1" {
			t.Fatalf("unexpected assignees: %+v", assignees)
		}

		// Users with tasks (should only be u1)
		withReq := httptest.NewRequest(http.MethodGet, "/api/users/with-tasks", nil)
		withRec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(withRec, withReq)
		var withTasks []db.User
		_ = json.NewDecoder(withRec.Body).Decode(&withTasks)
		if len(withTasks) != 1 || withTasks[0].ID != u1.ID {
			t.Fatalf("expected only u1 with tasks, got %+v", withTasks)
		}

		// Users without tasks (should only be u2)
		withoutReq := httptest.NewRequest(http.MethodGet, "/api/users/without-tasks", nil)
		withoutRec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(withoutRec, withoutReq)
		var withoutTasks []db.User
		_ = json.NewDecoder(withoutRec.Body).Decode(&withoutTasks)
		if len(withoutTasks) != 1 || withoutTasks[0].ID != u2.ID {
			t.Fatalf("expected only u2 without tasks, got %+v", withoutTasks)
		}

		// Unassign and re-assign via URL params
		unassignURL := fmt.Sprintf("/api/tasks/%d/assignees/%d", task.ID, u1.ID)
		unassignReq := httptest.NewRequest(http.MethodDelete, unassignURL, nil)
		unassignRec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(unassignRec, unassignReq)
		if unassignRec.Code != http.StatusNoContent {
			t.Fatalf("expected 204 on unassign user, got %d", unassignRec.Code)
		}

		// Re-assign for cascade delete test
		req2 := httptest.NewRequest(http.MethodPost, "/api/tasks/assign", bytes.NewBufferString(payload))
		rec2 := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(rec2, req2)
	})

	t.Run("deleting task cascades and frees user assignment", func(t *testing.T) {
		delReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/tasks/%d", task.ID), nil)
		delRec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(delRec, delReq)
		if delRec.Code != http.StatusNoContent {
			t.Fatalf("expected 204 on task delete, got %d", delRec.Code)
		}

		// Now users with tasks must be empty []
		withReq := httptest.NewRequest(http.MethodGet, "/api/users/with-tasks", nil)
		withRec := httptest.NewRecorder()
		env.server.Routes().ServeHTTP(withRec, withReq)
		if strings.TrimSpace(withRec.Body.String()) != "[]" {
			t.Fatalf("expected empty array [] after cascade delete, got %s", withRec.Body.String())
		}
	})
}
