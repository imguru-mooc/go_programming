package main

/*
extern void goCallback(int);
void run_with_callback(void);
*/
import "C"

import "fmt"

//export goCallback
func goCallback(n C.int) {
	fmt.Println("Go에서 호출됨:", int(n))
}

func main() {
	C.run_with_callback()
}
