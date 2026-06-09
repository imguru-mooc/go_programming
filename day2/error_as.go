package main

import (
	"errors"
	"fmt"
)

type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Msg)
}

func main() {
	var err error = fmt.Errorf("처리 실패: %w",
		&ValidationError{Field: "email", Msg: "형식 오류"})

	fmt.Println(err)
	var ve *ValidationError
	if errors.As(err, &ve) {
		fmt.Println("문제 필드:", ve.Field)
		fmt.Println("이유:", ve.Msg)
	}
}
