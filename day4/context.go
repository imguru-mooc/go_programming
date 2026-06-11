package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := serviceA(ctx)
	if err != nil {
		fmt.Println("최종 에러:", err)
		return
	}

	fmt.Println("작업 성공")
}

func serviceA(ctx context.Context) error {
	fmt.Println("serviceA 시작")
	return serviceB(ctx)
}

func serviceB(ctx context.Context) error {
	fmt.Println("  serviceB 시작")
	return repositoryC(ctx)
}

func repositoryC(ctx context.Context) error {
	fmt.Println("    repositoryC 시작")
	return externalAPID(ctx)
}

func externalAPID(ctx context.Context) error {
	fmt.Println("      externalAPID 시작")

	for i := 1; i <= 10; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
			fmt.Println("      외부 API 처리 중...", i)
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}
