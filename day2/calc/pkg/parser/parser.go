package parser

import (
	"fmt"
	"strconv"
)

// Args는 파싱된 명령 인자를 담는다
type Args struct {
	Op string
	A  float64
	B  float64
}

// Parse는 ["add", "3", "5"] 같은 슬라이스를 받아 Args로 변환한다
func Parse(args []string) (*Args, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("인자 3개가 필요합니다 (받은 개수: %d)", len(args))
	}

	a, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return nil, fmt.Errorf("첫 번째 피연산자 파싱 실패 (%q): %w", args[1], err)
	}

	b, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return nil, fmt.Errorf("두 번째 피연산자 파싱 실패 (%q): %w", args[2], err)
	}

	return &Args{Op: args[0], A: a, B: b}, nil
}
