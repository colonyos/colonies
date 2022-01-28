package main

import (
	"syscall/js"

	"github.com/colonyos/colonies/pkg/security/crypto"
)

func prvkey() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		prvKey, err := crypto.CreateCrypto().GeneratePrivateKey()
		if err != nil {
			return err.Error()
		}

		return prvKey
	})

	return jsonFunc
}

func id() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		prvKey := args[0].String()
		id, err := crypto.CreateCrypto().GenerateID(prvKey)
		if err != nil {
			return err.Error()
		}

		return id
	})

	return jsonFunc
}

func sign() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		msg := args[0].String()
		prvKey := args[1].String()
		signature, err := crypto.CreateCrypto().GenerateSignature(msg, prvKey)
		if err != nil {
			return err.Error()
		}

		return signature
	})

	return jsonFunc
}

func hash() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		msg := args[0].String()
		return crypto.CreateCrypto().GenerateHash(msg)
	})

	return jsonFunc
}

func recoverid() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		msg := args[0].String()
		signature := args[1].String()
		id, err := crypto.CreateCrypto().RecoverID(msg, signature)
		if err != nil {
			return err.Error()
		}

		return id
	})

	return jsonFunc
}

func main() {
	js.Global().Set("__cryptolib__prvkey", prvkey())
	js.Global().Set("__cryptolib__id", id())
	js.Global().Set("__cryptolib__sign", sign())
	js.Global().Set("__cryptolib__hash", hash())
	js.Global().Set("__cryptolib__recoverid", recoverid())

	<-make(chan bool)
}
