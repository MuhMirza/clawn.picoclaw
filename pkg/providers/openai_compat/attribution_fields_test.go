package openai_compat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func captureChatRequestBody(t *testing.T, apiBase string, options map[string]any) map[string]any {
	t.Helper()
	var requestBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp := map[string]any{
			"choices": []map[string]any{{
				"message": map[string]any{"content": "ok"},
				"finish_reason": "stop",
			}},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewProvider("key", server.URL, "")
	p.apiBase = apiBase
	p.httpClient = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			r.URL, _ = url.Parse(server.URL + r.URL.Path)
			return http.DefaultTransport.RoundTrip(r)
		}),
	}

	_, err := p.Chat(
		context.Background(),
		[]Message{{Role: "user", Content: "hi"}},
		nil,
		"gemini-3-flash-preview",
		options,
	)
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	return requestBody
}

func TestProviderChat_OmitsMetadataAndRequestTagsForGeminiDirect(t *testing.T) {
	body := captureChatRequestBody(t, "https://generativelanguage.googleapis.com/v1beta", map[string]any{
		"metadata": map[string]any{"clawn_agent_id": "main"},
		"request_tags": []string{"clawn-managed", "agent:main"},
	})
	if _, exists := body["metadata"]; exists {
		t.Fatalf("metadata should NOT be sent to Gemini direct API, got %#v", body)
	}
	if _, exists := body["request_tags"]; exists {
		t.Fatalf("request_tags should NOT be sent to Gemini direct API, got %#v", body)
	}
}

func TestProviderChat_KeepsMetadataAndRequestTagsForClawnManagedCompat(t *testing.T) {
	body := captureChatRequestBody(t, "http://litellm-proxy:4000/v1", map[string]any{
		"metadata": map[string]any{"clawn_agent_id": "main"},
		"request_tags": []string{"clawn-managed", "agent:main"},
	})
	if _, exists := body["metadata"]; !exists {
		t.Fatalf("metadata should be preserved for clawn-managed compat API, got %#v", body)
	}
	if _, exists := body["request_tags"]; !exists {
		t.Fatalf("request_tags should be preserved for clawn-managed compat API, got %#v", body)
	}
}
