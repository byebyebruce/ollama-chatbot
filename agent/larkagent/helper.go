package larkagent

import (
	"encoding/json"
	"strings"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func MessageText(event *larkim.P2MessageReceiveV1) (string, bool) {
	if event.Event.Message.MessageType == nil {
		return "", false
	}
	var (
		a       map[string]any
		content = *event.Event.Message.Content
	)
	if *event.Event.Message.MessageType == larkim.MsgTypeText {
		if err := json.Unmarshal([]byte(content), &a); err != nil {
			return "", false
		}
		str := a[larkim.MsgTypeText].(string)
		for _, m := range event.Event.Message.Mentions {
			str = strings.Replace(str, *m.Key, "", 1)
		}
		return str, true
	}
	return "", false
}

func IsGroupAtMessage(event *larkim.P2MessageReceiveV1, id string) bool {
	if event.Event.Message.ChatType == nil {
		return false
	}
	if *event.Event.Message.ChatType != "group" {
		return false
	}
	if len(event.Event.Message.Mentions) == 0 {
		return false
	}
	for _, m := range event.Event.Message.Mentions {
		if m.Id == nil {
			continue
		}

		if *m.Id.OpenId == id || *m.Id.UnionId == id {
			return true
		}
	}
	return false
}

// IsSingleChatMessage 是否单聊
func IsSingleChatMessage(event *larkim.P2MessageReceiveV1) bool {
	if event.Event.Message.ChatType == nil {
		return false
	}
	if *event.Event.Message.ChatType != "p2p" {
		return false
	}
	return true
}
