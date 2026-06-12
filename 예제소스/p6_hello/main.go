package main

/*
#include <stdio.h>
#include <stdlib.h>

void say_hello(const char *name) {
    printf("Hello from C, %s!\n", name);
    fflush(stdout);
}
*/
import "C"

import "unsafe"

func main() {
	name := C.CString("Go")
	defer C.free(unsafe.Pointer(name))
	C.say_hello(name)
}
