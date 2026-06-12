package main

/*
#cgo LDFLAGS: -lm
#include <math.h>
*/
import "C"

import "fmt"

func main() {
	result := C.sqrt(C.double(2.0))
	fmt.Println(float64(result))
}
