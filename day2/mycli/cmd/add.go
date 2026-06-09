package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [a] [b]",
	Short: "두 정수를 더합니다",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("첫 번째 값이 정수가 아닙니다: %w", err)
		}

		b, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("두 번째 값이 정수가 아닙니다: %w", err)
		}

		fmt.Println(a + b)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
