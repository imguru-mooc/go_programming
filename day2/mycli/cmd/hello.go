package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var upper bool

var helloCmd = &cobra.Command{
	Use:   "hello [name]",
	Short: "인사 메시지를 출력합니다",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if verbose {
			fmt.Println("[debug] hello command 실행")
		}

		if upper {
			name = strings.ToUpper(name)
		}

		fmt.Println("Hello,", name)
	},
}

func init() {
	rootCmd.AddCommand(helloCmd)

	helloCmd.Flags().BoolVarP(
		&upper,
		"upper",
		"u",
		false,
		"이름을 대문자로 출력합니다",
	)
}
