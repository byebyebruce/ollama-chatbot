package model

import (
	"bytes"
	"text/template"
)

// ChatMessage 聊天消息
type ChatMessage struct {
	MsgID      string                 // 消息唯一ID
	ThreadID   string                 // 会话ID(使用者自己抽象，可以是群组id，可以是用户id，一个会话id对应一个机器人)
	SenderID   string                 // 发送者ID
	SenderName string                 // 发送者名称
	ToUserID   string                 // 接收者ID
	ToUserName string                 // 接收到名称
	Content    string                 // 消息内容
	Meta       map[string]interface{} // 元数据
}

type MessageRender struct {
	tmpl *template.Template
}

func NewMessageRender(tmpl string) (*MessageRender, error) {
	if len(tmpl) == 0 {
		return &MessageRender{}, nil
	}
	t, err := template.New("msg").Parse(tmpl)
	if err != nil {
		return nil, err
	}
	return &MessageRender{tmpl: t}, nil
}
func (m MessageRender) Render(msg ChatMessage) string {
	if m.tmpl == nil {
		return msg.Content
	}
	b := &bytes.Buffer{}
	err := m.tmpl.Execute(b, msg)
	if err != nil {
		return msg.Content
	}
	return b.String()
}
