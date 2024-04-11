package cmd

import (
	"github.com/byebyebruce/ollamax"
	"github.com/ollama/ollama/format"
	"github.com/spf13/cobra"
)

func ListCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list all models",
		RunE: func(cmd *cobra.Command, args []string) error {
			models, err := ollamax.ListModel()
			if err != nil {
				return err
			}
			for _, m := range models {
				cmd.Println(m.Name, format.HumanBytes(m.Size))
			}
			return nil
		},
	}
}
