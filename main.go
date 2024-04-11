package main

import (
	"github.com/byebyebruce/ollama-chatbot/cmd"
)

func main() {
	rootCmd := cmd.RunWechatBotCmd()
	rootCmd.AddCommand(cmd.RunLarkBotCmd(), cmd.ListCMD(), cmd.PullCMD())
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
