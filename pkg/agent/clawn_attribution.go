package agent

import (
	"fmt"
	"strings"
)

func applyClawnAttribution(out map[string]any, agentID, model, channel, chatID string, metadataIn map[string]string) {
	if out == nil {
		return
	}

	metadata := map[string]any{
		"clawn_agent_id":       agentID,
		"clawn_agent_slug":     agentID,
		"clawn_runtime_engine": "picoclaw",
		"clawn_model_route":    model,
	}
	requestTags := []string{"clawn-managed", "runtime:picoclaw"}
	if agentID != "" {
		requestTags = append(requestTags, fmt.Sprintf("agent:%s", agentID))
	}
	if model != "" {
		requestTags = append(requestTags, fmt.Sprintf("route:%s", model))
	}
	if channel != "" {
		metadata["clawn_channel"] = channel
	}
	if chatID != "" {
		metadata["clawn_chat_id"] = chatID
	}
	if metadataIn != nil {
		if v := strings.TrimSpace(metadataIn["sender_id"]); v != "" {
			metadata["clawn_user_id"] = v
			out["user"] = "aiagenz"
		}
		if v := strings.TrimSpace(metadataIn["account_id"]); v != "" {
			metadata["clawn_account_id"] = v
		}
		if v := strings.TrimSpace(metadataIn["project_id"]); v != "" {
			metadata["clawn_project_id"] = v
			out["end_user"] = v
			requestTags = append(requestTags, fmt.Sprintf("project:%s", v))
		}
		if v := strings.TrimSpace(metadataIn["channel"]); v != "" && channel == "" {
			metadata["clawn_channel"] = v
		}
		if v := strings.TrimSpace(metadataIn["chat_id"]); v != "" && chatID == "" {
			metadata["clawn_chat_id"] = v
		}
	}

	out["metadata"] = metadata
	out["request_tags"] = requestTags
	out["agent_id"] = agentID
}
