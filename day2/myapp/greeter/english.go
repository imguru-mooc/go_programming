package greeter

// 외부에 노출됨 (대문자)
func Hello() string {
	return greeting() + ", world!"
}

// 패키지 내부 전용 (소문자)
func greeting() string {
	return "Hello"
}

// 구조체 필드도 같은 규칙
type Config struct {
	Name   string // 공개
	secret string // 비공개
}
