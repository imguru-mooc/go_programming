package main

import (
	"fmt"
	"unicode"
)

func isPalindrome(s string) bool {
	runes := []rune(s)
	i, j := 0, len(runes)-1
	for i < j {
		// 공백/구두점 건너뛰기
		for i < j && !isAlphaNum(runes[i]) {
			i++
		}
		for i < j && !isAlphaNum(runes[j]) {
			j--
		}
		if unicode.ToLower(runes[i]) != unicode.ToLower(runes[j]) {
			return false
		}
		i++
		j--
	}
	return true
}

func isAlphaNum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func main() {
	cases := []string{
		"racecar",
		"hello",
		"A man a plan a canal Panama",
		"기러기",
		"토마토",
		"안녕하세요",
		"Was it a car or a cat I saw?",
		"",
	}
	for _, c := range cases {
		fmt.Printf("%-40q → %v\n", c, isPalindrome(c))
	}
}
