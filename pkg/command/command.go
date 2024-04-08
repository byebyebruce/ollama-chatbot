package command

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/byebyebruce/ollama-chatbot/model"
	"github.com/samber/lo"
)

var (
	CmdPrefix = "/" // 命令前缀
	cmdHelp   = "help"
)

type Handler func(msg model.ChatMessage, param ...string) (string, error)

type Command struct {
	CMD     string  `json:"cmd"`
	Desc    string  `json:"desc"`
	Handler Handler `json:"-"`
}

type Commands struct {
	cmds map[string]*Command
}

func NewCommands(cmds ...*Command) (*Commands, error) {
	c := &Commands{
		cmds: map[string]*Command{},
	}
	for _, cmd := range cmds {
		err := c.RegisterCommand(cmd)
		if err != nil {
			return nil, err
		}
	}
	c.RegisterCommand(&Command{CMD: cmdHelp, Desc: "帮助", Handler: c.helpHandler})
	return c, nil
}

func (c *Commands) RegisterCommand(cmd *Command) error {
	if cmd.CMD == "" {
		return fmt.Errorf("cmd is empty")
	}
	if _, ok := c.cmds[cmd.CMD]; ok {
		return fmt.Errorf("cmd %s already registered", cmd.CMD)
	}
	c.cmds[cmd.CMD] = cmd
	return nil
}

func (c *Commands) ProcessCommand(msg model.ChatMessage) (string, error, bool) {
	name := msg.SenderName
	content := msg.Content

	if !strings.HasPrefix(content, CmdPrefix) {
		return "", nil, false
	}

	content = strings.TrimPrefix(content, CmdPrefix)
	cmds := strings.Split(content, " ")
	cmd, ok := c.cmds[strings.ToLower(cmds[0])]
	if !ok {
		return "", nil, false
	}
	log.Println("admin", name, ":", content)
	content = strings.TrimPrefix(content, cmd.CMD)
	msg.Content = content
	ret, err := cmd.Handler(msg, cmds[1:]...)
	return ret, err, true
}

func (c *Commands) helpHandler(msg model.ChatMessage, param ...string) (string, error) {
	cmds := lo.Values(c.cmds)
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].CMD < cmds[j].CMD
	})
	strs := make([]string, 0, len(cmds))
	for i, cmd := range cmds {
		strs = append(strs, fmt.Sprintf("%d. %s %s", i, cmd.CMD, cmd.Desc))
	}
	return strings.Join(strs, "\n"), nil
}
