package security

type Crypto interface {
	GeneratePrivateKey() (string, error)
	GenerateID(prvKey string) (string, error)
	GenerateSignature(jsonString string, prvKey string) (string, error)
	GenerateHash(data string) string
	RecoverID(jsonString string, signature string) (string, error)
}
