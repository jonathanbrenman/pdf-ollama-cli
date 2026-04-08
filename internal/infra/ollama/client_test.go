package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Generate(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		expectedResponse := "This is a summary"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST method, got %s", r.Method)
			}
			if r.URL.Path != "/api/generate" {
				t.Errorf("expected /api/generate path, got %s", r.URL.Path)
			}

			var req generateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}

			if req.Prompt == "" {
				t.Error("expected non-empty prompt")
			}

			resp := generateResponse{Response: expectedResponse}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(server.URL, "llama2", nil)
		got, err := client.Generate(context.Background(), "test prompt")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != expectedResponse {
			t.Errorf("expected %q, got %q", expectedResponse, got)
		}
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("internal error"))
		}))
		defer server.Close()

		client := NewClient(server.URL, "llama2", nil)
		_, err := client.Generate(context.Background(), "test prompt")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("empty response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := generateResponse{Response: ""}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(server.URL, "llama2", nil)
		_, err := client.Generate(context.Background(), "test prompt")

		if err == nil {
			t.Fatal("expected error due to empty response, got nil")
		}
	})

	t.Run("invalid json response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := NewClient(server.URL, "llama2", nil)
		_, err := client.Generate(context.Background(), "test prompt")

		if err == nil {
			t.Fatal("expected error due to invalid json, got nil")
		}
	})

	t.Run("empty endpoint", func(t *testing.T) {
		client := NewClient("", "llama2", nil)
		if client.endpoint == "" {
			t.Error("expected default endpoint from config, got empty")
		}
	})

	t.Run("invalid url error", func(t *testing.T) {
		client := NewClient("http://[::1]:80%2", "llama2", nil) // invalid escape in host
		_, err := client.Generate(context.Background(), "test")
		if err == nil {
			t.Fatal("expected error from NewRequest, got nil")
		}
	})
}
