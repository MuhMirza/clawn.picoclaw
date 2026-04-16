package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
)

// WebchatChannel implements a native PicoClaw channel for the Webchat UI.
// It exposes an HTTP endpoint that the Go Bridge backend connects to via SSE.
type WebchatChannel struct {
	*BaseChannel
	config  config.WebchatConfig
	server  *http.Server
	clients map[string]chan string // chatID -> chan for SSE chunks
	mu      sync.RWMutex
}

func NewWebchatChannel(cfg config.WebchatConfig, messageBus *bus.MessageBus) (*WebchatChannel, error) {
	base := NewBaseChannel("webchat", cfg, messageBus, cfg.AllowFrom)

	return &WebchatChannel{
		BaseChannel: base,
		config:      cfg,
		clients:     make(map[string]chan string),
	}, nil
}

func (c *WebchatChannel) Start(ctx context.Context) error {
	logger.InfoC("webchat", "Starting Webchat HTTP server...")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/webchat/stream", c.handleStream)

	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	c.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.ErrorCF("webchat", "Webchat server error", map[string]any{"error": err.Error()})
		}
	}()

	c.setRunning(true)
	logger.InfoCF("webchat", "Webchat channel listening on http://%s", map[string]any{"addr": addr})

	go func() {
		<-ctx.Done()
		c.Stop(context.Background())
	}()

	return nil
}

func (c *WebchatChannel) Stop(ctx context.Context) error {
	logger.InfoC("webchat", "Stopping Webchat HTTP server...")
	c.setRunning(false)

	if c.server != nil {
		c.server.Shutdown(ctx)
	}

	c.mu.Lock()
	for _, ch := range c.clients {
		close(ch)
	}
	c.clients = make(map[string]chan string)
	c.mu.Unlock()

	return nil
}

// Send implements Channel interface for outbound replies from AgentLoop
func (c *WebchatChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	if !c.IsRunning() {
		return fmt.Errorf("webchat server not running")
	}

	c.mu.RLock()
	clientChan, exists := c.clients[msg.ChatID]
	c.mu.RUnlock()

	if !exists {
		// Frontend disconnected before reply arrived
		logger.WarnCF("webchat", "No active SSE connection for reply", map[string]any{
			"chat_id": msg.ChatID,
		})
		return nil
	}

	// Send chunk to active SSE connection
	select {
	case clientChan <- msg.Content:
		return nil
	case <-time.After(2 * time.Second):
		return fmt.Errorf("timeout writing to client chunk channel")
	}
}

type webchatRequest struct {
	SessionID   string            `json:"session_id"`
	Content     string            `json:"content"`
	User        string            `json:"user"`
	EndUser     string            `json:"end_user"`
	AgentID     string            `json:"agent_id"`
	Metadata    map[string]any    `json:"metadata"`
	RequestTags []string          `json:"request_tags"`
}

// handleStream handles inbound POST requests and streams the reply back as SSE
func (c *WebchatChannel) handleStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req webchatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Setup SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Ensure flush support
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	clientID := req.SessionID
	if clientID == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}

	// Set up receive channel for this client request
	ch := make(chan string, 100)
	c.mu.Lock()
	c.clients[clientID] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.clients, clientID)
		close(ch)
		c.mu.Unlock()
	}()

	metadata := map[string]string{
		"peer_kind": "direct",
		"peer_id":   clientID,
		"user_id":   clientID,
	}
	if req.User != "" {
		metadata["sender_id"] = req.User
		metadata["user_id"] = req.User
	}
	if req.EndUser != "" {
		metadata["project_id"] = req.EndUser
	}
	if req.AgentID != "" {
		metadata["agent_id"] = req.AgentID
	}
	if len(req.RequestTags) > 0 {
		if tagsJSON, err := json.Marshal(req.RequestTags); err == nil {
			metadata["request_tags"] = string(tagsJSON)
		}
	}
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			switch val := v.(type) {
			case string:
				if val != "" {
					metadata[k] = val
				}
			case fmt.Stringer:
				metadata[k] = val.String()
			case float64:
				metadata[k] = fmt.Sprintf("%v", val)
			case bool:
				metadata[k] = fmt.Sprintf("%v", val)
			default:
				if encoded, err := json.Marshal(v); err == nil {
					metadata[k] = string(encoded)
				}
			}
		}
	}

	// Fire into PicoClaw bus
	c.HandleMessage(
		clientID,
		clientID,
		req.Content,
		nil,
		metadata,
	)

	// Stream responses back to caller
	timeout := time.After(3 * time.Minute) // Failsafe timeout
	for {
		select {
		case <-r.Context().Done():
			// Client disconnected
			return
		case <-timeout:
			// Safety timeout
			return
		case content, ok := <-ch:
			if !ok {
				// Channel closed
				return
			}

			// Build OpenAI-compatible chunk format (since frontend expects SSE from our bridge)
			chunk := map[string]any{
				"choices": []map[string]any{
					{
						"delta": map[string]string{
							"content": content,
						},
					},
				},
			}

			chunkJSON, _ := json.Marshal(chunk)
			fmt.Fprintf(w, "data: %s\n\n", chunkJSON)
			flusher.Flush()

			// If it's the final chunk of a reply sequence (PicoClaw sends it all at once currently)
			// we send [DONE] to close the stream gracefully in the bridge.
			fmt.Fprintf(w, "data: [DONE]\n\n")
			flusher.Flush()
			return // Webchat replies are single-turn fire-and-forget for now
		}
	}
}
