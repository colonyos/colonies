package security

type Crypto interface {
	GeneratePrivateKey() (string, error)
	GenerateID(prvKey string) (string, error)
	GenerateSignature(data string, prvKey string) (string, error)
	GenerateHash(data string) string
	RecoverID(data string, signature string) (string, error)
}
