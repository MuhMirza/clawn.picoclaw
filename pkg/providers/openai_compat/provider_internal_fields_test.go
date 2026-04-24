package openai_compat

import "testing"

func TestSanitizeProviderMetadata_StripsInternalFieldsForThirdPartyHosts(t *testing.T) {
	meta := map[string]any{
		"clawn_agent_id":                 "main",
		"clawn_project_id":               "proj-1",
		"preflightReserveId":             "reserve-1",
		"managedLLMDecisionMode":         "allow_reserved",
		"managedLLMAvailableUnitsBefore": int64(123),
		"preflightReservedUnits":         int64(1),
		"project_id":                     "proj-1",
		"sender_id":                      "user-1",
		"chat_id":                        "chat-1",
		"safe_user_metadata":             "ok",
	}

	got := sanitizeProviderMetadata("https://generativelanguage.googleapis.com/v1beta/openai", meta)
	if got == nil {
		t.Fatalf("expected non-nil filtered metadata")
	}
	if _, exists := got["clawn_agent_id"]; exists {
		t.Fatalf("did not expect clawn_agent_id for third-party host: %#v", got)
	}
	if _, exists := got["preflightReserveId"]; exists {
		t.Fatalf("did not expect preflightReserveId for third-party host: %#v", got)
	}
	if _, exists := got["project_id"]; exists {
		t.Fatalf("did not expect project_id for third-party host: %#v", got)
	}
	if got["safe_user_metadata"] != "ok" {
		t.Fatalf("expected safe metadata to survive, got %#v", got)
	}
}

func TestSanitizeProviderRequestTags_StripsInternalTagsForThirdPartyHosts(t *testing.T) {
	tags := []string{"clawn-managed", "agent:main", "project:proj-1", "runtime:picoclaw", "route:gemini/gemini-3-flash-preview", "safe-tag"}
	got := sanitizeProviderRequestTags("https://generativelanguage.googleapis.com/v1beta/openai", tags)
	if len(got) != 1 || got[0] != "safe-tag" {
		t.Fatalf("expected only safe-tag to survive, got %#v", got)
	}
}

func TestSanitizeProviderMetadata_PreservesInternalFieldsForManagedHosts(t *testing.T) {
	meta := map[string]any{
		"clawn_agent_id":     "main",
		"clawn_project_id":   "proj-1",
		"preflightReserveId": "reserve-1",
	}
	got := sanitizeProviderMetadata("https://ai.clawn.id/v1", meta)
	if got == nil {
		t.Fatalf("expected metadata for managed-compatible host")
	}
	if got["clawn_agent_id"] != "main" {
		t.Fatalf("expected clawn_agent_id preserved, got %#v", got)
	}
	if got["preflightReserveId"] != "reserve-1" {
		t.Fatalf("expected preflightReserveId preserved, got %#v", got)
	}
}

func TestSanitizeProviderRequestTags_PreservesInternalTagsForManagedHosts(t *testing.T) {
	tags := []string{"clawn-managed", "agent:main", "project:proj-1", "runtime:picoclaw", "route:gemini/gemini-3-flash-preview"}
	got := sanitizeProviderRequestTags("https://ai.clawn.id/v1", tags)
	if len(got) != len(tags) {
		t.Fatalf("expected managed tags preserved, got %#v", got)
	}
}
