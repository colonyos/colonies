package core

import (
	"encoding/json"
	"strings"
	"time"
)

// FileEventKind discriminates the three observable transitions on a
// ColonyFS file. Wire-stable integer values: do not renumber.
type FileEventKind int

const (
	// FileAdded fires the first time a (colony, label, name) tuple is
	// written. Subsequent writes of the same name produce FileUpdated.
	FileAdded FileEventKind = 1
	// FileUpdated fires when an existing (colony, label, name) gets a
	// new revision. Subscribers that only care about novelty filter on
	// FileAdded alone.
	FileUpdated FileEventKind = 2
	// FileRemoved fires after RemoveFileByID or RemoveFileByName.
	FileRemoved FileEventKind = 3
)

// String returns the lowercase kind name. Useful for logs and CLI output;
// the integer value is the wire representation.
func (k FileEventKind) String() string {
	switch k {
	case FileAdded:
		return "added"
	case FileUpdated:
		return "updated"
	case FileRemoved:
		return "removed"
	}
	return "unknown"
}

// FileEvent is delivered to subscribers over the realtime websocket on
// every observable file mutation. Marshals to JSON for wire transport.
//
// FileID, Size, and Checksum are only populated for FileAdded and
// FileUpdated — for FileRemoved the file is gone, so the event identifies
// it by (ColonyName, Label, Name) only.
type FileEvent struct {
	Kind        FileEventKind `json:"kind"`
	ColonyName  string        `json:"colonyname"`
	Label       string        `json:"label"`
	Name        string        `json:"name"`
	FileID      string        `json:"fileid,omitempty"`
	Size        int64         `json:"size,omitempty"`
	Checksum    string        `json:"checksum,omitempty"`
	ChecksumAlg string        `json:"checksumalg,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
}

// MatchesPrefix reports whether the event's Label is matched by a
// subscription's labelPrefix.
//
// Rules:
//   - Empty prefix matches every label (whole-colony subscription).
//   - Exact match on the prefix passes.
//   - The event's label starting with prefix + "/" passes (any descendant
//     label inherits the subscription).
//
// Regular files at exactly `<prefix>/foo.md` are matched via the descendant
// rule. The function does NOT match `<prefix>foo` (no slash) — labels are
// path-shaped strings and a missing separator means a different label.
func (e *FileEvent) MatchesPrefix(prefix string) bool {
	if prefix == "" {
		return true
	}
	if e.Label == prefix {
		return true
	}
	return strings.HasPrefix(e.Label, prefix+"/")
}

// MatchesKinds reports whether the event passes the kind filter.
// An empty or nil filter matches every kind ("all kinds" sentinel).
func (e *FileEvent) MatchesKinds(kinds []int) bool {
	if len(kinds) == 0 {
		return true
	}
	for _, k := range kinds {
		if FileEventKind(k) == e.Kind {
			return true
		}
	}
	return false
}

// ToJSON marshals the event to its wire form. The Timestamp uses the
// default RFC3339 encoding via Go's time.Time JSON marshaller.
func (e *FileEvent) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// CreateFileEventFromJSON inverts ToJSON. Returns the event and any
// unmarshal error encountered.
func CreateFileEventFromJSON(jsonString string) (*FileEvent, error) {
	var ev *FileEvent
	if err := json.Unmarshal([]byte(jsonString), &ev); err != nil {
		return nil, err
	}
	return ev, nil
}

// CreateFileAddedEvent builds an event for the first-revision case.
func CreateFileAddedEvent(file *File, ts time.Time) *FileEvent {
	return &FileEvent{
		Kind:        FileAdded,
		ColonyName:  file.ColonyName,
		Label:       file.Label,
		Name:        file.Name,
		FileID:      file.ID,
		Size:        file.Size,
		Checksum:    file.Checksum,
		ChecksumAlg: file.ChecksumAlg,
		Timestamp:   ts,
	}
}

// CreateFileUpdatedEvent builds an event for a new revision of an
// existing (colony, label, name) tuple.
func CreateFileUpdatedEvent(file *File, ts time.Time) *FileEvent {
	return &FileEvent{
		Kind:        FileUpdated,
		ColonyName:  file.ColonyName,
		Label:       file.Label,
		Name:        file.Name,
		FileID:      file.ID,
		Size:        file.Size,
		Checksum:    file.Checksum,
		ChecksumAlg: file.ChecksumAlg,
		Timestamp:   ts,
	}
}

// CreateFileRemovedEvent builds an event for a removal. The file may
// already be gone from storage by the time the caller invokes this, so
// the caller passes the identifying coordinates explicitly rather than
// a *File pointer.
func CreateFileRemovedEvent(colonyName, label, name string, ts time.Time) *FileEvent {
	return &FileEvent{
		Kind:       FileRemoved,
		ColonyName: colonyName,
		Label:      label,
		Name:       name,
		Timestamp:  ts,
	}
}

// Equals deep-compares two events. Useful in tests; not used on the
// server-side fan-out path.
func (e *FileEvent) Equals(o *FileEvent) bool {
	if o == nil {
		return false
	}
	return e.Kind == o.Kind &&
		e.ColonyName == o.ColonyName &&
		e.Label == o.Label &&
		e.Name == o.Name &&
		e.FileID == o.FileID &&
		e.Size == o.Size &&
		e.Checksum == o.Checksum &&
		e.ChecksumAlg == o.ChecksumAlg &&
		e.Timestamp.Equal(o.Timestamp)
}
