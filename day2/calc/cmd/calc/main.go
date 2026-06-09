package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/myname/calc/internal/calculator"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:     "calc",
	Short:   "간단한 계산기",
	Long:    "사칙연산과 나머지 연산을 지원하는 CLI 계산기입니다.",
	Version: Version,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "상세 출력")

	rootCmd.AddCommand(makeOpCmd("add", "더하기", "+"))
	rootCmd.AddCommand(makeOpCmd("sub", "빼기", "-"))
	rootCmd.AddCommand(makeOpCmd("mul", "곱하기", "×"))
	rootCmd.AddCommand(makeOpCmd("div", "나누기", "÷"))
	rootCmd.AddCommand(makeOpCmd("mod", "나머지", "mod"))
}

func makeOpCmd(op, desc, sym string) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s [a] [b]", op),
		Short: desc,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return fmt.Errorf("a 파싱 실패: %w", err)
			}
			b, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return fmt.Errorf("b 파싱 실패: %w", err)
			}

			if verbose {
				fmt.Printf("[%s] %v %s %v 계산 중...\n", op, a, sym, b)
			}

			result, err := calculator.Calculate(op, a, b)
			if err != nil {
				return err
			}

			fmt.Printf("%g %s %g = %g\n", a, sym, b, result)
			return nil
		},
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
