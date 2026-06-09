package main

import "fmt"

// Go — 관리된 방식
func mayPanic() {
	panic("심각한 오류!")
}

func safeCall() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("패닉 복구:", r)
		}
	}()
	mayPanic()
	fmt.Println("이 라인은 실행 안 됨")
}

func main() {
	safeCall()
	fmt.Println("프로그램 계속")
}
