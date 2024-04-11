package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/byebyebruce/ollama-chatbot/agent/wechatagent"
	"github.com/byebyebruce/ollama-chatbot/chat"
	"github.com/byebyebruce/ollama-chatbot/pkg/configloader"
	"github.com/byebyebruce/ollama-chatbot/pkg/persist/boltdb"
	"github.com/byebyebruce/ollamax"
	"github.com/spf13/cobra"
)

type WechatBotConfig struct {
	WeChatAgentConfig wechatagent.WeChatAgentConfig `json:"wechat" yaml:"wechat"` // 聊天配置
	LLM               chat.ChatConfig               `json:"llm" yaml:"llm"`       // 机器人配置
}

// RunWechatBotCmd 个人微信聊天机器人
func RunWechatBotCmd() *cobra.Command {
	var (
		configPath     string
		loginCachePath string
		db             string
	)
	c := &cobra.Command{
		Short: "个人微信聊天机器人，扫码登陆后，将群组改成配置里的后缀(如：-AIChat)",
	}
	c.Flags().StringVarP(&configPath, "config", "c", "config.yaml", "config file path")
	c.Flags().StringVarP(&loginCachePath, "wx", "w", "login_cache.json", "wx storage file path")
	c.Flags().StringVarP(&db, "db", "d", "wechat_bot.db", "db file")
	c.RunE = func(cmd *cobra.Command, args []string) error {
		cfg := &WechatBotConfig{}
		err := configloader.Load(configPath, cfg)
		if err != nil {
			panic(err)
		}

		p, err := boltdb.NewBoltDB(db, "wechat_bot")
		if err != nil {
			panic(err)
		}

		if err := ollamax.Init(); err != nil {
			panic(err)
		}
		defer ollamax.Cleanup()

		chat, err := chat.NewChat(cfg.LLM, p)
		if err != nil {
			panic(err)
		}
		defer chat.Close()

		ctx, cancel := context.WithCancel(context.Background())

		wb := wechatagent.New(cfg.WeChatAgentConfig, chat)
		go func() {
			defer cancel()
			err = wb.Run(loginCachePath)
			if err != nil {
				log.Println(err)
			}
		}()

		// 创建一个通道来接收操作系统信号
		sigChan := make(chan os.Signal, 1)
		// 通知信号处理程序捕获 SIGINT（Ctrl+C）
		signal.Notify(sigChan, syscall.SIGINT)
		select {
		case <-ctx.Done():
		case <-sigChan: // 阻塞直到收到 SIGINT
		}
		fmt.Println("准备退出...")
		return nil
	}

	return c
}
