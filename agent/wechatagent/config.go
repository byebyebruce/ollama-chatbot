package wechatagent

import (
	"strings"
)

type WeChatAgentConfig struct {
	GroupSuffix   string `json:"group_suffix" yaml:"group_suffix"`     // 群组名后缀，只有以这些后缀结尾的群组才会开启
	MsgTmpl       string `json:"msg_tmpl" yaml:"msg_tmpl"`             // 消息模板
	TriggerPrefix string `json:"trigger_prefix" yaml:"trigger_prefix"` // 聊天过滤前缀
	FilterPrefix  string `json:"filter_prefix" yaml:"filter_prefix"`   // 聊天过滤前缀
}

func (c WeChatAgentConfig) MatchGroup(str string) bool {
	return strings.HasSuffix(str, c.GroupSuffix)
}

func (c WeChatAgentConfig) ShouldTrigger(text string) (string, bool) {
	// 优先判断触发
	if len(c.TriggerPrefix) > 0 {
		if strings.HasPrefix(text, c.TriggerPrefix) {
			return strings.TrimPrefix(text, c.TriggerPrefix), true
		}
		return "", false
	}
	// 再判断是否需要过滤
	if len(c.FilterPrefix) > 0 {
		if strings.HasPrefix(text, c.FilterPrefix) {
			return "", false
		}
	}
	return text, true
}
