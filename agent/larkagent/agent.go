package larkagent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/byebyebruce/ollama-chatbot/admin"
	"github.com/byebyebruce/ollama-chatbot/chat"
	"github.com/byebyebruce/ollama-chatbot/model"

	"github.com/gin-gonic/gin"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/patrickmn/go-cache"
)

type LarkAgentConfig struct {
	AppID             string `yaml:"app_id"`
	AppSecret         string `yaml:"app_secret"`
	VerificationToken string `yaml:"verification_token"`
	EncryptKey        string `yaml:"encrypt_key"`
}

type LarkAgent struct {
	*lark.Client
	cfg   LarkAgentConfig
	chat  chat.ChatInterface
	admin *admin.Admin
	block *cache.Cache
}

func New(cfg LarkAgentConfig, chat *chat.Chat) *LarkAgent {
	larkClient := lark.NewClient(cfg.AppID, cfg.AppSecret)

	admin := admin.NewAdmin(chat)
	return &LarkAgent{
		Client: larkClient,
		cfg:    cfg,
		chat:   chat,
		admin:  admin,
		block:  cache.New(time.Second*30, time.Second*10),
	}
}

func (w *LarkAgent) Run(addr string, path string) error {
	router := gin.Default()

	// 注册消息处理器
	// https://github.com/larksuite/oapi-sdk-go/blob/v3_main/README.zh.md#%E9%9B%86%E6%88%90gin%E6%A1%86%E6%9E%B6
	dispatcher := dispatcher.NewEventDispatcher(w.cfg.VerificationToken, w.cfg.EncryptKey)
	dispatcher.OnP2MessageReceiveV1(w.onReceiveMessage)

	handler := func(ctx *gin.Context) {
		httpserverext.NewEventHandlerFunc(dispatcher)(ctx.Writer, ctx.Request)
	}
	router.POST(path, handler)

	return router.Run(addr)
}

func (b *LarkAgent) onReceiveMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	var (
		//chatID = ""
		//senderUnionID = *event.Event.Sender.SenderId.UnionId
		senderID = *event.Event.Sender.SenderId.UserId
		//senderOpenID = *event.Event.Sender.SenderId.OpenId
		msgID = *event.Event.Message.MessageId
	)

	_, ok := b.block.Get(senderID)
	if ok {
		// duplicate message
		return b.ReplyTextMessage(ctx, msgID, "发送太频繁")
	}
	b.block.Add(senderID, nil, time.Minute*2)

	switch {
	//case IsGroupAtMessage(event, b.botID): // 群聊 at
	//	chatID = *event.Event.Message.ChatId
	case IsSingleChatMessage(event): // 单聊
	default:
		return nil
	}

	text, ok := MessageText(event)
	if !ok {
		return nil
	}

	cm := model.ChatMessage{
		MsgID:      msgID,
		ThreadID:   senderID,
		SenderID:   senderID,
		SenderName: senderID,
		//ToUserID:   msg.ToUserName,
		//ToUserName: msg.ToNickName,
		Content: text,
	}
	// 处理管理员消息, 优先级最高
	if ret, err, ok := b.admin.ProcessCommand(cm); ok {
		log.Println("Admin:\n", ret, err)
		if err != nil {
			return b.ReplyTextMessage(ctx, msgID, "Admin: error\n"+err.Error())
		} else {
			return b.ReplyTextMessage(ctx, msgID, "Admin: "+ret)
		}
	}

	go func() {
		defer b.block.Delete(senderID)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
		defer cancel()

		text, err := b.chat.Chat(ctx, cm)
		if err != nil {
			text = fmt.Sprintf("error: %s", err.Error())
			log.Println("chat", err)
		}
		if err != nil {
			b.ReplyTextMessage(ctx, msgID, "error\n"+err.Error())
		} else {
			b.ReplyTextMessage(ctx, msgID, text)
		}
	}()
	return nil
}

func (f *LarkAgent) ReplyTextMessage(ctx context.Context, msgID string, text string) error {
	req := larkim.NewReplyMessageReqBuilder().
		MessageId(msgID).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			Content(BuildTextMessage(text)).
			MsgType(larkim.MsgTypeText).
			ReplyInThread(false).
			//Uuid("a0d69e20-1dd1-458b-k525-dfeca4015204").
			Build()).
		Build()
	// 发起请求
	resp, err := f.Client.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		return err
	}

	// 服务端错误处理
	if !resp.Success() {
		return fmt.Errorf("code:%v,msg:%v,req:%v",
			resp.Code, resp.Msg, resp.RequestId())
	}
	return nil
}

func BuildAtUserMessage(id, content string) string {
	c := fmt.Sprintf(`<at user_id="%s"></at>`, id)
	return buildMessage(larkim.MsgTypeText, c+content)
}

func BuildAtAllMessage(content string) string {
	c := fmt.Sprintf(`<at user_id="all"></at>%s`, content)
	return buildMessage(larkim.MsgTypeText, c)
}

func BuildTextMessage(content string) string {
	return buildMessage(larkim.MsgTypeText, content)
}

func buildMessage(t, content string) string {
	b, _ := json.Marshal(map[string]interface{}{
		t: content,
	})
	return string(b)
}
