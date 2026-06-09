package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/myname/calc/internal/calculator"
	"github.com/myname/calc/pkg/parser"
	"path/filepath"
	"time"
)

var Version = "dev"

func appendHistory(line string) error {
	if os.Getenv("CALC_HISTORY") != "on" {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".calc-history")

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	ts := time.Now().Format("2006-01-02 15:04:05")
	_, err = fmt.Fprintf(f, "%s | %s\n", ts, line)
	return err
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	// --version 처리
	if os.Args[1] == "--version" || os.Args[1] == "-v" {
		fmt.Printf("calc %s\n", Version)
		return
	}

	args, err := parser.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "에러:", err)
		os.Exit(1)
	}

	result, err := calculator.Calculate(args.Op, args.A, args.B)
	if err != nil {
		if errors.Is(err, calculator.ErrDivByZero) {
			fmt.Fprintln(os.Stderr, "❌ 0으로 나눌 수 없습니다")
		} else {
			fmt.Fprintln(os.Stderr, "에러:", err)
		}
		os.Exit(1)
	}

	fmt.Printf("%g %s %g = %g\n", args.A, symbol(args.Op), args.B, result)

	line := fmt.Sprintf("%g %s %g = %g", args.A, symbol(args.Op), args.B, result)

	if err := appendHistory(line); err != nil {
		fmt.Fprintln(os.Stderr, "이력 저장 실패:", err)
	}
}

func symbol(op string) string {
	switch op {
	case "add":
		return "+"
	case "sub":
		return "-"
	case "mul":
		return "×"
	case "div":
		return "÷"
	case "mod":
		return "%"
	}
	return op
}

func usage() {
	fmt.Println("사용법: calc <op> <a> <b>")
	fmt.Println("  op: add | sub | mul | div")
	fmt.Println("예시: calc add 3 5")
	fmt.Println()
	fmt.Println("옵션:")
	fmt.Println("  -v, --version  버전 표시")
}
