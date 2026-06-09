package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "mycli",
	Short: "mycli는 Cobra 예제 CLI입니다",
	Long:  "mycli는 Go Cobra를 단계적으로 배우기 위한 예제 프로그램입니다.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mycli 실행됨")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"자세한 로그를 출력합니다",
	)
}
