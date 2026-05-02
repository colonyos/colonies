package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Round-trip a fully-populated SubscribeFilesMsg through JSON and back.
func TestRPCSubscribeFilesMsg(t *testing.T) {
	msg := CreateSubscribeFilesMsg("test_colony", "/home/root/inbox", []int{1, 2}, 30)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateSubscribeFilesMsgFromJSON(jsonString + "garbage")
	assert.NotNil(t, err)

	msg2, err = CreateSubscribeFilesMsgFromJSON(jsonString)
	assert.Nil(t, err)
	assert.True(t, msg.Equals(msg2))
}

// Indented form must round-trip identically; we use it in CLI debug output.
func TestRPCSubscribeFilesMsgIndent(t *testing.T) {
	msg := CreateSubscribeFilesMsg("test_colony", "/home/root/inbox", []int{1}, 0)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateSubscribeFilesMsgFromJSON(jsonString + "garbage")
	assert.NotNil(t, err)

	msg2, err = CreateSubscribeFilesMsgFromJSON(jsonString)
	assert.Nil(t, err)
	assert.True(t, msg.Equals(msg2))
}

// Equals must compare every field including the Kinds slice element-wise.
// Tests every field as the discriminator one at a time, plus nil and the
// kinds-length / kinds-content corner cases that a naive shallow compare
// would miss.
func TestRPCSubscribeFilesMsgEquals(t *testing.T) {
	base := CreateSubscribeFilesMsg("c1", "/p", []int{1, 2}, 30)

	// nil and self
	assert.False(t, base.Equals(nil))
	assert.True(t, base.Equals(base))

	// Each field as the only difference
	differs := []*SubscribeFilesMsg{
		CreateSubscribeFilesMsg("c2", "/p", []int{1, 2}, 30),       // colony
		CreateSubscribeFilesMsg("c1", "/other", []int{1, 2}, 30),   // prefix
		CreateSubscribeFilesMsg("c1", "/p", []int{1}, 30),          // kinds length
		CreateSubscribeFilesMsg("c1", "/p", []int{1, 3}, 30),       // kinds content
		CreateSubscribeFilesMsg("c1", "/p", []int{1, 2}, 60),       // timeout
	}
	for i, d := range differs {
		assert.False(t, base.Equals(d), "differs[%d] should not be equal", i)
	}

	// MsgType drift — matters because servers reject mismatches as a
	// defense against payload-type confusion.
	odd := CreateSubscribeFilesMsg("c1", "/p", []int{1, 2}, 30)
	odd.MsgType = "unrelated"
	assert.False(t, base.Equals(odd))

	// Identical Kinds slices but different backing arrays must still equal.
	a := CreateSubscribeFilesMsg("c1", "/p", []int{1, 2}, 30)
	b := CreateSubscribeFilesMsg("c1", "/p", append([]int{}, 1, 2), 30)
	assert.True(t, a.Equals(b))
}

// Empty kinds is the documented "all kinds" sentinel; round-trip preserves
// the empty-slice (or nil) shape so the server-side matcher can handle it.
func TestRPCSubscribeFilesMsgEmptyKinds(t *testing.T) {
	msg := CreateSubscribeFilesMsg("c1", "/p", nil, 30)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateSubscribeFilesMsgFromJSON(jsonString)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(msg2.Kinds))
	assert.True(t, msg.Equals(msg2))

	// Explicit empty slice should also round-trip and equal nil.
	msgEmpty := CreateSubscribeFilesMsg("c1", "/p", []int{}, 30)
	assert.True(t, msg.Equals(msgEmpty))
}

// MsgType is set by the constructor and used for routing on the server. If
// this constant ever drifts, every existing client breaks silently — pin it
// here so a regression shows up immediately.
func TestRPCSubscribeFilesMsgPayloadType(t *testing.T) {
	assert.Equal(t, "subscribefilesmsg", SubscribeFilesPayloadType)
	msg := CreateSubscribeFilesMsg("c1", "/p", nil, 0)
	assert.Equal(t, SubscribeFilesPayloadType, msg.MsgType)
}

// JSON tag names are part of the wire contract. Older clients/servers
// would silently misparse if a tag is renamed, so guard the on-the-wire
// shape explicitly.
func TestRPCSubscribeFilesMsgJSONShape(t *testing.T) {
	msg := CreateSubscribeFilesMsg("c1", "/home/root", []int{1, 3}, 60)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)
	for _, want := range []string{
		`"colonyname":"c1"`,
		`"labelprefix":"/home/root"`,
		`"kinds":[1,3]`,
		`"timeout":60`,
		`"msgtype":"subscribefilesmsg"`,
	} {
		assert.Contains(t, jsonString, want)
	}
}
