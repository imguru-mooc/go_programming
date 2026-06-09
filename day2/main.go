// main.go
package main

import "fmt"

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	fmt.Printf("v%s built at %s\n", Version, BuildTime)
}
