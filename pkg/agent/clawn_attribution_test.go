package agent

import "testing"

func TestApplyClawnAttribution_DoesNotExposeTopLevelAgentID(t *testing.T) {
	out := map[string]any{}
	applyClawnAttribution(out, "main", "gemini/gemini-3-flash-preview", "pico", "chat-1", map[string]string{
		"sender_id":  "user-1",
		"project_id": "proj-1",
	})

	if _, exists := out["agent_id"]; exists {
		t.Fatalf("did not expect top-level agent_id in generic llm attribution payload, got %#v", out)
	}

	metadata, _ := out["metadata"].(map[string]any)
	if metadata == nil {
		t.Fatalf("expected metadata map, got %#v", out)
	}
	if got, _ := metadata["clawn_agent_id"].(string); got != "main" {
		t.Fatalf("expected clawn_agent_id metadata preserved, got %#v", metadata)
	}

	tags, _ := out["request_tags"].([]string)
	foundAgentTag := false
	for _, tag := range tags {
		if tag == "agent:main" {
			foundAgentTag = true
			break
		}
	}
	if !foundAgentTag {
		t.Fatalf("expected agent request tag preserved, got %#v", tags)
	}
}
