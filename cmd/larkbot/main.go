// Package 个人微信聊天机器人
package main

import (
	"context"
	"flag"
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
)

var (
	configPath = flag.String("config", "config.yaml", "config file path")
	db         = flag.String("db", "lark_bot.db", "db file")
	addr       = flag.String("addr", ":8080", "server address")
	urlPath    = flag.String("url-path", "/", "server url path")
)

type Config struct {
	LarkAgentConfig larkagent.LarkAgentConfig `json:"lark" yaml:"lark"` // 聊天配置
	Chat            chat.ChatConfig           `json:"chat" yaml:"chat"` // 机器人配置
}

func main() {
	flag.Parse()

	cfg := &Config{}
	err := configloader.Load(*configPath, cfg)
	if err != nil {
		panic(err)
	}

	p, err := boltdb.NewBoltDB(*db, "lark_bot")
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

		err = wb.Run(*addr, *urlPath)
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
}
