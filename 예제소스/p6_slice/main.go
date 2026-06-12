// 6.6 — Go 슬라이스를 C 배열로 전달 (전체 코드)
package main

/*
#include <stddef.h>
#include <stdio.h>

// 받은 바이트들의 합을 계산하고 내용을 출력하는 C 함수
unsigned long process_bytes(const unsigned char *data, size_t len) {
    unsigned long sum = 0;
    printf("C가 받은 데이터: ");
    for (size_t i = 0; i < len; i++) {
        printf("%d ", data[i]);
        sum += data[i];
    }
    printf("\n");
    fflush(stdout);
    return sum;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func processBytes(data []byte) uint64 {
	if len(data) == 0 { // ⚠️ 빈 슬라이스에서 &data[0]은 panic
		return 0
	}
	sum := C.process_bytes(
		(*C.uchar)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
	)
	return uint64(sum)
}

func main() {
	data := []byte{1, 2, 3, 4, 5}
	fmt.Println("합계(C에서 계산):", processBytes(data))
	fmt.Println("빈 슬라이스:", processBytes(nil)) // 안전하게 0
}
