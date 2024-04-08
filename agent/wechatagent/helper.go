package wechatagent

import (
	"fmt"
	"strings"

	"github.com/eatmoreapple/openwechat"
)

// ProcessSystemMessage process system message
func ProcessSystemMessage(msg *openwechat.Message) (bool, error) {
	if !msg.IsSystem() {
		return false, nil
	}
	u, err := msg.Sender()
	if err != nil {
		return true, err
	}
	g, ok := u.AsGroup()
	if !ok {
		return true, nil
	}
	return true, g.Detail()
}

// ParseGroupAndSender return group name and sender name
func ParseGroupAndSender(msg *openwechat.Message) (senderID, senderName, groupID, groupName string, retErr error) {
	var (
		sender    *openwechat.User
		groupUser *openwechat.User
	)
	groupUser, err := msg.Sender()
	if err != nil {
		retErr = err
		return
	}
	if groupUser.IsSelf() {
		sender = groupUser
		groupUser, err = msg.Receiver()
		if err != nil {
			retErr = err
			return
		}
	} else {
		sender, err = msg.SenderInGroup()
		if err != nil {
			retErr = err
			return
		}
	}

	group, ok := groupUser.AsGroup()
	if !ok {
		retErr = fmt.Errorf("not group")
		return
	}
	if len(groupUser.NickName) == 0 {
		if err = groupUser.Detail(); err != nil {
			retErr = err
			return
		}
	}

	senderID = sender.UserName
	senderName = sender.NickName
	groupID = group.UserName
	groupName = group.NickName

	return
}

// IsQuote 判断是否是引用消息
func IsQuote(msg string) bool {
	prefix := "「"
	sufix := "」"
	if strings.HasPrefix(msg, prefix) && strings.HasSuffix(msg, sufix) {
		return true
	}
	return false
}
