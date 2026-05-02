package core

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// FileEventKind values are wire-stable. If anyone renumbers, every
// connected client misroutes — pin the integers.
func TestFileEventKindWireValues(t *testing.T) {
	assert.Equal(t, FileEventKind(1), FileAdded)
	assert.Equal(t, FileEventKind(2), FileUpdated)
	assert.Equal(t, FileEventKind(3), FileRemoved)
}

func TestFileEventKindString(t *testing.T) {
	assert.Equal(t, "added", FileAdded.String())
	assert.Equal(t, "updated", FileUpdated.String())
	assert.Equal(t, "removed", FileRemoved.String())
	assert.Equal(t, "unknown", FileEventKind(42).String())
}

func TestFileEventJSONRoundTrip(t *testing.T) {
	ts := time.Date(2026, 5, 2, 14, 30, 0, 0, time.UTC)
	ev := &FileEvent{
		Kind:        FileAdded,
		ColonyName:  "ai",
		Label:       "/home/root/inbox",
		Name:        "morning-brief.md",
		FileID:      "abc123",
		Size:        4096,
		Checksum:    "deadbeef",
		ChecksumAlg: "sha256",
		Timestamp:   ts,
	}

	jsonString, err := ev.ToJSON()
	assert.Nil(t, err)
	for _, want := range []string{
		`"kind":1`,
		`"colonyname":"ai"`,
		`"label":"/home/root/inbox"`,
		`"name":"morning-brief.md"`,
		`"fileid":"abc123"`,
		`"size":4096`,
		`"checksum":"deadbeef"`,
		`"checksumalg":"sha256"`,
	} {
		assert.Contains(t, jsonString, want)
	}

	parsed, err := CreateFileEventFromJSON(jsonString)
	assert.Nil(t, err)
	assert.True(t, ev.Equals(parsed))
}

// FileRemoved events legitimately omit FileID/Size/Checksum (the file is
// already gone). The omitempty tags must keep them out of the JSON so
// receivers don't see a misleading FileID="" or Size=0.
func TestFileRemovedEventOmitsImmutableFields(t *testing.T) {
	ts := time.Date(2026, 5, 2, 14, 30, 0, 0, time.UTC)
	ev := CreateFileRemovedEvent("ai", "/home/root/inbox", "old.md", ts)

	jsonString, err := ev.ToJSON()
	assert.Nil(t, err)
	assert.NotContains(t, jsonString, "fileid")
	assert.NotContains(t, jsonString, "size")
	assert.NotContains(t, jsonString, "checksum")
	// Required fields still present.
	assert.Contains(t, jsonString, `"kind":3`)
	assert.Contains(t, jsonString, `"colonyname":"ai"`)
	assert.Contains(t, jsonString, `"name":"old.md"`)
}

func TestFileEventMatchesPrefix(t *testing.T) {
	ev := &FileEvent{Label: "/home/root/inbox"}

	// Empty prefix is the whole-colony wildcard.
	assert.True(t, ev.MatchesPrefix(""))
	// Exact match.
	assert.True(t, ev.MatchesPrefix("/home/root/inbox"))
	// Parent label — descendant rule.
	assert.True(t, ev.MatchesPrefix("/home/root"))
	assert.True(t, ev.MatchesPrefix("/home"))
	// Sibling label — must not match.
	assert.False(t, ev.MatchesPrefix("/home/root/outbox"))
	// String prefix without separator must not match. /home/root/in is
	// a different label from /home/root/inbox; a naive HasPrefix would
	// pass, the path-shaped rule must not.
	assert.False(t, ev.MatchesPrefix("/home/root/in"))
	// A descendant event matches its ancestor prefix but the converse
	// must not.
	deeper := &FileEvent{Label: "/home/root/inbox/2026"}
	assert.True(t, deeper.MatchesPrefix("/home/root/inbox"))
	assert.False(t, ev.MatchesPrefix("/home/root/inbox/2026"))
}

func TestFileEventMatchesKinds(t *testing.T) {
	added := &FileEvent{Kind: FileAdded}
	updated := &FileEvent{Kind: FileUpdated}
	removed := &FileEvent{Kind: FileRemoved}

	// Empty / nil filter is the all-kinds sentinel.
	assert.True(t, added.MatchesKinds(nil))
	assert.True(t, added.MatchesKinds([]int{}))

	// Single-kind filters
	assert.True(t, added.MatchesKinds([]int{int(FileAdded)}))
	assert.False(t, added.MatchesKinds([]int{int(FileUpdated)}))

	// Multi-kind filters
	assert.True(t, removed.MatchesKinds([]int{int(FileAdded), int(FileRemoved)}))
	assert.False(t, updated.MatchesKinds([]int{int(FileAdded), int(FileRemoved)}))

	// Garbage kind values in the filter are tolerated; they just don't
	// match anything.
	assert.False(t, added.MatchesKinds([]int{99}))
}

func TestCreateFileAddedEvent(t *testing.T) {
	ts := time.Date(2026, 5, 2, 14, 30, 0, 0, time.UTC)
	file := &File{
		ID:          "abc",
		ColonyName:  "ai",
		Label:       "/home/root/inbox",
		Name:        "x.md",
		Size:        128,
		Checksum:    "cs",
		ChecksumAlg: "sha256",
	}
	ev := CreateFileAddedEvent(file, ts)
	assert.Equal(t, FileAdded, ev.Kind)
	assert.Equal(t, "ai", ev.ColonyName)
	assert.Equal(t, "/home/root/inbox", ev.Label)
	assert.Equal(t, "x.md", ev.Name)
	assert.Equal(t, "abc", ev.FileID)
	assert.Equal(t, int64(128), ev.Size)
	assert.Equal(t, "cs", ev.Checksum)
	assert.Equal(t, "sha256", ev.ChecksumAlg)
	assert.Equal(t, ts, ev.Timestamp)
}

func TestCreateFileUpdatedEvent(t *testing.T) {
	file := &File{ID: "id", ColonyName: "c", Label: "/l", Name: "n", Size: 1, Checksum: "x"}
	ts := time.Now()
	ev := CreateFileUpdatedEvent(file, ts)
	assert.Equal(t, FileUpdated, ev.Kind)
	assert.Equal(t, "id", ev.FileID)
}

func TestCreateFileRemovedEvent(t *testing.T) {
	ts := time.Now()
	ev := CreateFileRemovedEvent("c", "/l", "n", ts)
	assert.Equal(t, FileRemoved, ev.Kind)
	assert.Equal(t, "c", ev.ColonyName)
	assert.Equal(t, "", ev.FileID)
	assert.Equal(t, int64(0), ev.Size)
}

func TestFileEventEquals(t *testing.T) {
	ts := time.Now()
	a := CreateFileRemovedEvent("c", "/l", "n", ts)

	// nil and self
	assert.False(t, a.Equals(nil))
	assert.True(t, a.Equals(a))

	// Different field flips equality each time
	b := CreateFileRemovedEvent("c", "/l", "n", ts)
	assert.True(t, a.Equals(b))
	b.Kind = FileAdded
	assert.False(t, a.Equals(b))

	c := CreateFileRemovedEvent("c", "/l", "n", ts)
	c.ColonyName = "other"
	assert.False(t, a.Equals(c))

	d := CreateFileRemovedEvent("c", "/l", "n", ts)
	d.Timestamp = ts.Add(time.Second)
	assert.False(t, a.Equals(d))
}

// JSON tag names are part of the wire contract; older clients silently
// misparse if a tag is renamed. Pin every tag here.
func TestFileEventJSONShape(t *testing.T) {
	ev := CreateFileAddedEvent(&File{
		ID:          "id",
		ColonyName:  "c",
		Label:       "/l",
		Name:        "n",
		Size:        7,
		Checksum:    "cs",
		ChecksumAlg: "sha256",
	}, time.Now())

	raw, err := json.Marshal(ev)
	assert.Nil(t, err)

	var into map[string]interface{}
	assert.Nil(t, json.Unmarshal(raw, &into))

	for _, want := range []string{"kind", "colonyname", "label", "name", "fileid", "size", "checksum", "checksumalg", "timestamp"} {
		_, ok := into[want]
		assert.True(t, ok, "expected field %s in marshalled JSON: %s", want, string(raw))
	}
}
