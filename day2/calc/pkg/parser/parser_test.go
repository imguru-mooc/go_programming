package parser

import "testing"

func TestParse_Valid(t *testing.T) {
	args, err := Parse([]string{"add", "3", "5"})
	if err != nil {
		t.Fatalf("예상치 못한 에러: %v", err)
	}
	if args.Op != "add" || args.A != 3 || args.B != 5 {
		t.Errorf("파싱 결과 불일치: %+v", args)
	}
}

func TestParse_WrongCount(t *testing.T) {
	_, err := Parse([]string{"add", "3"})
	if err == nil {
		t.Error("에러가 발생해야 함")
	}
}

func TestParse_InvalidNumber(t *testing.T) {
	_, err := Parse([]string{"add", "abc", "5"})
	if err == nil {
		t.Error("에러가 발생해야 함")
	}
}
