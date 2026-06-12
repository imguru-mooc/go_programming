package main

/*
#include "myadd.h"
*/
import "C"

import "fmt"

func main() {
	result := C.my_add(C.int(3), C.int(4))
	fmt.Println("결과:", int(result))
}
