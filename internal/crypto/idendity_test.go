package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateIdendity(t *testing.T) {
	idendity, err := CreateIdendityFromString("6d2fb6f546bacfd98c68769e61e0b44a697a30596c018a50e28200aa59b01c0a")
	assert.Nil(t, err)

	assert.Equal(t, "4fef2b5a82d134d058c1883c72d6d9caf77cd59ca82d73105017590dea3dcb87", idendity.ID())
	assert.Equal(t, "6d2fb6f546bacfd98c68769e61e0b44a697a30596c018a50e28200aa59b01c0a", idendity.PrivateKeyAsHex())
	assert.Equal(t, "0408e903276ee7973666dceeefa5335e5c4b6b5989821906db98f8de8acf8f853824ca3234a8602200baa2d75f30cb2050cda18602824c3eb2da654a93a01a7ad4", idendity.PublicKeyAsHex())
}
