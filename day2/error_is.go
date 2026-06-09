package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {
	_, err := os.Open("sample.txt")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("파일이 없습니다")
		} else if errors.Is(err, os.ErrPermission) {
			fmt.Println(os.ErrPermission)
			fmt.Println("권한 없음")
		} else {
			fmt.Println("기타 에러:", err)
		}
	}
}

/*
package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

func main() {
	err := fmt.Errorf("사용자 조회 실패: %w", ErrNotFound)

	fmt.Println(err)

	if errors.Is(err, ErrNotFound) {
		fmt.Println("원인은 ErrNotFound입니다")
	}
}
*/
