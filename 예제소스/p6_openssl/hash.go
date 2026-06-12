package main

/*
#cgo pkg-config: openssl
#include <openssl/sha.h>
#include <string.h>

void compute_sha256(const unsigned char *data, size_t len, unsigned char *digest) {
    SHA256(data, len, digest);
}
*/
import "C"

import (
	"encoding/hex"
	"fmt"
	"unsafe"
)

func sha256OpenSSL(data []byte) string {
	digest := make([]byte, 32)
	if len(data) == 0 {
		C.compute_sha256(nil, 0, (*C.uchar)(unsafe.Pointer(&digest[0])))
	} else {
		C.compute_sha256(
			(*C.uchar)(unsafe.Pointer(&data[0])),
			C.size_t(len(data)),
			(*C.uchar)(unsafe.Pointer(&digest[0])),
		)
	}
	return hex.EncodeToString(digest)
}

func main() {
	hash := sha256OpenSSL([]byte("Hello, CGo!"))
	fmt.Println("SHA-256:", hash)
}
