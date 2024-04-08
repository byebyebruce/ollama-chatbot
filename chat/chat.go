package chat

import (
	"context"
	"fmt"

	"github.com/byebyebruce/ollama-chatbot/model"
	"github.com/byebyebruce/ollama-chatbot/pkg/persist"
	"github.com/byebyebruce/ollamax"
	"github.com/ollama/ollama/api"
)

type ChatConfig struct {
	SystemPrompt string `json:"system_prompt" yaml:"system_prompt"`
	Model        string `json:"model" yaml:"model"`
	MaxHistory   int    `json:"max_history" yaml:"max_history"`
}

func historyID(id string) string {
	return fmt.Sprintf("history-%s", id)
}

type ChatInterface interface {
	Chat(ctx context.Context, message model.ChatMessage) (string, error)
}

var _ ChatInterface = (*Chat)(nil)

type Chat struct {
	llms    *ollamax.Ollamax
	persist persist.Persistent
	ChatConfig
}

func NewChat(cfg ChatConfig, persist persist.Persistent) (*Chat, error) {
	if cfg.MaxHistory <= 0 {
		cfg.MaxHistory = 20
	}

	llms, err := ollamax.NewWithAutoDownload(cfg.Model)
	if err != nil {
		return nil, err
	}
	c := &Chat{
		llms:       llms,
		ChatConfig: cfg,
		persist:    persist,
	}
	return c, nil
}

func (c *Chat) ClearHistory(threadID string) error {
	historyID := historyID(threadID)
	return c.persist.Delete(historyID)
}

func (c *Chat) Chat(ctx context.Context, message model.ChatMessage) (string, error) {
	var (
		threadID = message.ThreadID
		content  = message.Content
	)
	// 历史
	var history []api.Message
	historyKey := historyID(threadID)
	if _, err := c.persist.Load(historyKey, &history); err != nil {
		return "", err
	} else {
		for len(history) > c.ChatConfig.MaxHistory {
			history = history[2:]
		}
	}

	var msg []api.Message
	if len(c.ChatConfig.SystemPrompt) > 0 {
		msg = append(msg, api.Message{Role: "system", Content: c.ChatConfig.SystemPrompt})
	}
	msg = append(msg, history...)
	msg = append(msg, api.Message{Role: "user", Content: content})

	resp, err := c.llms.ChatStream(ctx, msg)
	if err != nil {
		return "", err
	}

	fullText := ""
LOOP:
	for {
		select {
		case <-ctx.Done():
			return fullText, nil
		case response, ok := <-resp:
			if !ok {
				break LOOP
			}
			if response.Err != nil {
				return "", response.Err
			}
			fullText += response.Result.Content
		}
	}
	history = append(history, api.Message{Role: "user", Content: content})
	history = append(history, api.Message{Role: "assistant", Content: fullText})
	if err = c.persist.Save(historyKey, history); err != nil {
		return "", err
	}
	return fullText, nil
}

func (c *Chat) Close() {
	c.llms.Close()
}
