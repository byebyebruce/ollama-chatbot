package admin

import (
	"encoding/json"

	"github.com/byebyebruce/ollama-chatbot/chat"
	"github.com/byebyebruce/ollama-chatbot/model"
	"github.com/byebyebruce/ollama-chatbot/pkg/command"
)

const (
	cmdClear  = "clear"  // 清除历史
	cmdPrompt = "prompt" // 设置提示语
	cmdInfo   = "info"   // 显示信息
)

type Admin struct {
	m *chat.Chat
	*command.Commands
}

func NewAdmin(m *chat.Chat) *Admin {
	a := &Admin{
		m: m,
	}

	cmds := []*command.Command{
		{CMD: cmdClear, Desc: "清除聊天历史", Handler: a.clearHandler},
		//{CMD: cmdPrompt, Desc: "设置自定义prompt", Handler: a.promptHandler},
		{CMD: cmdInfo, Desc: "显示信息", Handler: a.infoHandler},
	}
	c, err := command.NewCommands(cmds...)
	if err != nil {
		panic(err)
	}
	a.Commands = c

	return a
}
func (c *Admin) clearHandler(msg model.ChatMessage, param ...string) (string, error) {
	err := c.m.ClearHistory(msg.ThreadID)
	if err != nil {
		return "", err
	}
	return "clear success", nil
}

func (c *Admin) infoHandler(msg model.ChatMessage, param ...string) (string, error) {
	bs, _ := json.MarshalIndent(c.m.ChatConfig, " ", " ")
	return string(bs), nil
}

/*

func (c *Admin) promptHandler(msg model.ChatMessage, param ...string) (string, error) {
	cfg, err := c.m.GetConfig(msg.ThreadID)
	if err != nil {
		return "", err
	}
	cfg.SystemPrompt = msg.Content
	err = c.m.SetConfig(msg.ThreadID, *cfg)
	if err != nil {
		return "", err
	}
	c.m.ClearHistory(msg.ThreadID)

	return "change ok", nil
}


func (c *Admin) historyHandler(msg model.ChatMessage, params ...string) (string, error) {
	if len(params) < 1 {
		return "", fmt.Errorf("param should be 0~N")
	}
	n, err := strconv.Atoi(params[0])
	if err != nil {
		return "", fmt.Errorf("param should be 0~N")
	}
	if n > 20 {
		n = 20
	}

	cfg, err := c.m.GetConfig(msg.ThreadID)
	if err != nil {
		return "", err
	}
	cfg.MaxHistory = n
	err = c.m.SetConfig(msg.ThreadID, *cfg)
	if err != nil {
		return "", err
	}
	c.m.ClearHistory(msg.ThreadID)
	return "ok", nil
}


*/
