package main

import (
	"C"
	"colonies/pkg/security/crypto"
	"fmt"
)

//export sign
func sign(cmsg *C.char, cprv *C.char) *C.char {
	msg := C.GoString(cmsg)
	prvKey := C.GoString(cprv)
	signature, err := crypto.CreateCrypto().GenerateSignature(msg, prvKey)
	if err != nil {
		fmt.Println("failed to generate signature")
		return C.CString("ERROR")
	}

	return C.CString(signature)
}

func main() {}
