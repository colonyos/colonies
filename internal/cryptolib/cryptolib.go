package main

import (
	"C"
	"colonies/pkg/crypto"
	"encoding/hex"
	"fmt"
)

//export sign
func sign(cmsg *C.char, cprv *C.char) *C.char {
	msg := C.GoString(cmsg)
	prvKey := C.GoString(cprv)
	idendity, err := crypto.CreateIdendityFromString(prvKey)
	hash := crypto.GenerateHashFromString(msg)
	signature, err := crypto.Sign(hash, idendity.PrivateKey())
	if err != nil {
		fmt.Println("failed to generate signature")
		return C.CString("ERROR")
	}

	signatureBytes := hex.EncodeToString(signature)

	return C.CString(string(signatureBytes))
}

func main() {}
