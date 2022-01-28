package main

import (
	"C"
	"fmt"

	"github.com/colonyos/colonies/pkg/security/crypto"
)

//export prvkey
func prvkey() *C.char {
	prvKey, err := crypto.CreateCrypto().GeneratePrivateKey()
	if err != nil {
		fmt.Println("failed to private key")
		return C.CString("error")
	}

	return C.CString(prvKey)
}

//export id
func id(cprvkey *C.char) *C.char {
	prvKey := C.GoString(cprvkey)
	id, err := crypto.CreateCrypto().GenerateID(prvKey)
	if err != nil {
		fmt.Println("failed to generate id from private key")
		return C.CString("error")
	}

	return C.CString(id)
}

//export sign
func sign(cmsg *C.char, cprvkey *C.char) *C.char {
	msg := C.GoString(cmsg)
	prvKey := C.GoString(cprvkey)
	signature, err := crypto.CreateCrypto().GenerateSignature(msg, prvKey)
	if err != nil {
		fmt.Println("failed to generate signature")
		return C.CString("error")
	}

	return C.CString(signature)
}

//export hash
func hash(cmsg *C.char) *C.char {
	msg := C.GoString(cmsg)
	hash := crypto.CreateCrypto().GenerateHash(msg)
	return C.CString(hash)
}

//export recoverid
func recoverid(cmsg *C.char, csig *C.char) *C.char {
	msg := C.GoString(cmsg)
	signature := C.GoString(csig)
	id, err := crypto.CreateCrypto().RecoverID(msg, signature)
	if err != nil {
		fmt.Println("failed to recover id")
		return C.CString("error")
	}

	return C.CString(id)
}

func main() {}
