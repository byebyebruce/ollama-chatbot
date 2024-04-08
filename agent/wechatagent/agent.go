// Package wechatagent 个人微信聊天机器人
package wechatagent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/byebyebruce/ollama-chatbot/admin"
	"github.com/byebyebruce/ollama-chatbot/chat"
	"github.com/byebyebruce/ollama-chatbot/model"
	"github.com/byebyebruce/ollama-chatbot/pkg/stringx"

	"github.com/eatmoreapple/openwechat"
	"github.com/patrickmn/go-cache"
)

const (
	// 最大回复字数
	maxReplyLen = 2048
)

// WeChatAgent 个人微信聊天机器人
type WeChatAgent struct {
	cfg    WeChatAgentConfig
	chat   *chat.Chat
	admin  *admin.Admin
	block  *cache.Cache
	render *model.MessageRender
}

func New(cfg WeChatAgentConfig, chat *chat.Chat) *WeChatAgent {
	admin := admin.NewAdmin(chat)
	render, err := model.NewMessageRender(cfg.MsgTmpl)
	if err != nil {
		panic(fmt.Errorf("parse config error %v", err))
	}
	return &WeChatAgent{
		cfg:    cfg,
		chat:   chat,
		admin:  admin,
		block:  cache.New(time.Second*30, time.Second*10),
		render: render,
	}
}

func (w *WeChatAgent) Run(storageFile string) error {
	wechatBot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式
	// 注册登陆二维码回调
	wechatBot.UUIDCallback = openwechat.PrintlnQrcodeUrl

	// 创建热存储容器对象
	reloadStorage := openwechat.NewFileHotReloadStorage(storageFile)
	defer reloadStorage.Close()

	// 执行热登录
	//if err := wechatBot.PushLogin(reloadStorage, openwechat.NewRetryLoginOption()); err != nil {
	if err := wechatBot.HotLogin(reloadStorage, openwechat.NewRetryLoginOption()); err != nil {
		return err
	}

	// 获取登陆的用户
	self, err := wechatBot.GetCurrentUser()
	if err != nil {
		return err
	}
	log.Println("login ok:", self.UserName, self.NickName, self.DisplayName)

	// 获取所有的好友
	fs, err := self.Friends(true)
	if err != nil {
		panic(err)
	}
	log.Println("total friends: ", len(fs))

	// 获取所有的群组
	groups, err := self.Groups(true)
	if err != nil {
		panic(err)
	}
	log.Println(groups, err)

	log.Println("start ok...")

	// 注册消息处理函数
	wechatBot.MessageHandler = w.OnMessage

	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	return wechatBot.Block()
}

func (w *WeChatAgent) OnMessage(msg *openwechat.Message) {
	//log.Println(msg.FromUserName, msg.ToUserName, msg.Content)

	if !msg.IsSendByGroup() {
		return
	}

	// 处理系统信息
	if ok, _ := ProcessSystemMessage(msg); ok {
		return
	}

	// 只处理文本消息
	if !msg.IsText() {
		return
	}

	// 不处理@消息
	if msg.IsAt() {
		return
	}

	if len(msg.Content) == 0 {
		return
	}

	// 不处理引用消息
	if IsQuote(msg.Content) {
		return
	}

	// 获取发送者和群组
	senderID, senderName, groupID, groupName, err := ParseGroupAndSender(msg)
	if err != nil {
		log.Println("ParseGroupAndSender", err)
		return
	}

	cm := model.ChatMessage{
		MsgID:      msg.MsgId,
		ThreadID:   groupID,
		SenderID:   senderID,
		SenderName: senderName,
		ToUserID:   msg.ToUserName,
		//ToUserName: msg.ToNickName,
		Content: msg.Content,
	}
	// 处理管理员消息, 优先级最高
	if ret, err, ok := w.admin.ProcessCommand(cm); ok {
		log.Println("Admin:\n", ret, err)
		if err != nil {
			msg.ReplyText("Admin: error\n" + err.Error())
		} else {
			msg.ReplyText("Admin:\n" + ret)
		}
		return
	}

	// 群组过滤
	if !w.cfg.MatchGroup(groupName) {
		return
	}

	// 判断是否触发
	if text, ok := w.cfg.ShouldTrigger(msg.Content); !ok {
		return
	} else {
		cm.Content = text
	}

	log.Println("chat:", groupName, cm.SenderName, cm.Content)

	go func() {
		sessID := fmt.Sprintf("%s-%s", cm.ThreadID, cm.SenderID)
		if err := w.block.Add(sessID, struct{}{}, time.Second*30); err != nil {
			msg.ReplyText("@" + cm.SenderName + "\n稍等，前面的问题还没有回答")
			return
		}
		defer w.block.Delete(sessID)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*40)
		defer cancel()

		text, err := w.chat.Chat(ctx, cm)
		if err != nil {
			text = fmt.Sprintf("error: %s", err.Error())
			log.Println("chat", err)
		}
		text = w.render.Render(model.ChatMessage{
			SenderID:   senderID,
			SenderName: senderName,
			Content:    text,
			ThreadID:   groupID,
		})
		if len(text) == 0 {
			return
		}

		text = stringx.TruncateString(text, maxReplyLen)
		_, err = msg.ReplyText(text)
		if err != nil {
			log.Println("Reply", err)
			return
		}
	}()
}
