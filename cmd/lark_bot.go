package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/byebyebruce/ollama-chatbot/agent/larkagent"
	"github.com/byebyebruce/ollama-chatbot/chat"
	"github.com/byebyebruce/ollama-chatbot/pkg/configloader"
	"github.com/byebyebruce/ollama-chatbot/pkg/persist/boltdb"
	"github.com/byebyebruce/ollamax"
	"github.com/spf13/cobra"
)

type Config struct {
	LarkAgentConfig larkagent.LarkAgentConfig `json:"lark" yaml:"lark"`
	Chat            chat.ChatConfig           `json:"llm" yaml:"llm"`
}

func RunLarkBotCmd() *cobra.Command {
	var (
		configPath string
		db         string
		addr       string
		urlPath    string
	)
	c := &cobra.Command{
		Use:   "lark_bot",
		Short: "飞书机器人",
	}
	c.Flags().StringVarP(&configPath, "config", "c", "config.yaml", "config file path")
	c.Flags().StringVarP(&db, "db", "d", "lark_bot.db", "db file")
	c.Flags().StringVarP(&addr, "addr", "a", ":8080", "server address")
	c.Flags().StringVarP(&urlPath, "url-path", "u", "/", "server url path")

	c.RunE = func(cmd *cobra.Command, args []string) error {

		cfg := &Config{}
		err := configloader.Load(configPath, cfg)
		if err != nil {
			panic(err)
		}

		p, err := boltdb.NewBoltDB(db, "lark_bot")
		if err != nil {
			panic(err)
		}

		if err := ollamax.Init(); err != nil {
			panic(err)
		}
		defer ollamax.Cleanup()

		chat, err := chat.NewChat(cfg.Chat, p)
		if err != nil {
			panic(err)
		}
		defer chat.Close()

		ctx, cancel := context.WithCancel(context.Background())
		wb := larkagent.New(cfg.LarkAgentConfig, chat)
		go func() {
			defer cancel()

			err = wb.Run(addr, urlPath)
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
