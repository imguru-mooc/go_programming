package main

import (
	"fmt"
	"os"
)

func readConfig(path string) error {
	_, err := os.Open(path)
	if err != nil {
		// %w로 원본 에러를 래핑
		return fmt.Errorf("설정 파일 열기 실패 (%s): %w", path, err)
	}
	return nil
}

func main() {
	err := readConfig("none.yaml")
	if err != nil {
		fmt.Println(err)
		// 출력:
		// 설정 파일 열기 실패 (none.yaml): open none.yaml: no such file or directory
	}
}
