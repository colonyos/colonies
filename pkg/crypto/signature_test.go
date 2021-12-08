package crypto

import (
	. "colonies/pkg/utils"
	"encoding/hex"
	"testing"
)

func TestRecoveredID(t *testing.T) {
	idendity, err := CreateIdendity()
	CheckError(t, err)

	msg := "test"
	hash := GenerateHashFromString(msg)

	sig, err := Sign(hash, idendity.PrivateKey())
	CheckError(t, err)

	recoveredID, err := RecoveredID(hash, sig)
	CheckError(t, err)

	if recoveredID != idendity.ID() {
		Fatal(t, "idendity and recovered id does not match")
	}
}

func TestRecoverFromStrings(t *testing.T) {
	idendity, err := CreateIdendity()
	CheckError(t, err)

	msg := "test"
	hash := GenerateHashFromString(msg)

	sig, err := Sign(hash, idendity.PrivateKey())
	CheckError(t, err)

	hash2, err := CreateHashFromString(hash.String())
	CheckError(t, err)

	recoveredID, err := RecoveredID(hash2, sig)
	CheckError(t, err)

	if recoveredID != idendity.ID() {
		Fatal(t, "recovered and ididenty id does not match")
	}
}

func TestRecoverFromStrings2(t *testing.T) {
	idendity, err := CreateIdendity()
	CheckError(t, err)

	msg := "test"
	hash := GenerateHashFromString(msg)

	sig, err := Sign(hash, idendity.PrivateKey())
	CheckError(t, err)

	hash = GenerateHash([]byte("blablabla"))

	recoveredID, err := RecoveredID(hash, sig)
	CheckError(t, err)

	if recoveredID == idendity.ID() {
		Fatal(t, "recovered idendity does not match")
	}
}

func TestRecoverPublicKey(t *testing.T) {
	idendity, err := CreateIdendity()
	CheckError(t, err)

	hash := GenerateHash([]byte("test"))

	sig, err := Sign(hash, idendity.PrivateKey())
	CheckError(t, err)

	pub, err := RecoverPublicKey(hash, sig)
	CheckError(t, err)

	if idendity.PublicKeyAsHex() != hex.EncodeToString(pub) {
		t.Fatalf("invalid recovered public key")
	}
}

func TestRecoverPublicKeyInvalidSignature(t *testing.T) {
	idendity, err := CreateIdendity()
	CheckError(t, err)

	idendity2, err := CreateIdendity()
	CheckError(t, err)

	hash := GenerateHash([]byte("test"))

	sig, err := Sign(hash, idendity2.PrivateKey())
	CheckError(t, err)

	pub, err := RecoverPublicKey(hash, sig)
	CheckError(t, err)

	if idendity.PublicKeyAsHex() == hex.EncodeToString(pub) {
		t.Fatalf("invalid recovered public key")
	}
}

func TestSignAndVerify(t *testing.T) {
	idendity, err := CreateIdendity()
	CheckError(t, err)

	msg := "test"
	hash := GenerateHashFromString(msg)

	sig, err := Sign(hash, idendity.PrivateKey())
	CheckError(t, err)

	decPub, err := hex.DecodeString(idendity.PublicKeyAsHex())
	CheckError(t, err)

	ok, err := Verify(decPub, hash, sig)
	CheckError(t, err)
	if !ok {
		Fatal(t, "invalid signature")
	}
}

func TestSignAndVerifyInvalidPubKey(t *testing.T) {
	idendity, err := CreateIdendity()
	CheckError(t, err)

	idendity2, err := CreateIdendity()
	CheckError(t, err)

	msg := "test"
	hash := GenerateHashFromString(msg)

	sig, err := Sign(hash, idendity.PrivateKey())
	CheckError(t, err)

	decPub, err := hex.DecodeString(idendity2.PublicKeyAsHex())
	CheckError(t, err)

	ok, err := Verify(decPub, hash, sig)
	CheckError(t, err)

	if ok {
		Fatal(t, "expected an invalid signature")
	}
}
