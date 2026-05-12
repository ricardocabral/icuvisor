package intervals

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTrainingPlanPreservesAssignmentRaw(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/athlete/i12345/training-plan"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"training_plan_id":456,"training_plan_start_date":"2026-02-01","alias":"Base","training_plan":{"name":"Base plan","children":[{"id":1},{"id":2}]}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	plan, err := client.GetTrainingPlan(context.Background())
	if err != nil {
		t.Fatalf("GetTrainingPlan() error = %v", err)
	}
	if !plan.Active {
		t.Fatal("Active = false, want true")
	}
	if plan.TrainingPlanID != "456" {
		t.Fatalf("TrainingPlanID = %q, want 456", plan.TrainingPlanID)
	}
	if plan.Raw["training_plan"] == nil {
		t.Fatal("Raw training_plan missing")
	}
}

func TestGetTrainingPlanNotFoundIsNoActivePlan(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	plan, err := client.GetTrainingPlan(context.Background())
	if err != nil {
		t.Fatalf("GetTrainingPlan() error = %v", err)
	}
	if plan.Active {
		t.Fatal("Active = true, want false for 404/no active plan")
	}
}

func TestGetTrainingPlanPropagatesUnauthorized(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "no", http.StatusUnauthorized)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	_, err := client.GetTrainingPlan(context.Background())
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("GetTrainingPlan() error = %v, want ErrUnauthorized", err)
	}
}
