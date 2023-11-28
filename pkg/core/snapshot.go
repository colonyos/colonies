package core

import (
	"encoding/json"
	"time"
)

type Snapshot struct {
	ID         string    `json:"snapshotid"`
	ColonyName string    `json:"colonyname"`
	Label      string    `json:"label"`
	Name       string    `json:"name"`
	FileIDs    []string  `json:"fileids"`
	Added      time.Time `json:"added"`
}

func ConvertJSONToSnapshot(jsonString string) (*Snapshot, error) {
	var snapshot *Snapshot
	err := json.Unmarshal([]byte(jsonString), &snapshot)
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (snapshot *Snapshot) Equals(snapshot2 *Snapshot) bool {
	same := true

	if snapshot.ID != snapshot2.ID {
		same = false
	}
	if snapshot.ColonyName != snapshot2.ColonyName {
		same = false
	}
	if snapshot.Label != snapshot2.Label {
		same = false
	}
	if snapshot.Name != snapshot2.Name {
		same = false
	}
	if snapshot.Added.Unix() != snapshot2.Added.Unix() {
		same = false
	}

	if len(snapshot.FileIDs) != len(snapshot2.FileIDs) {
		return false
	}

	for i := range snapshot.FileIDs {
		if snapshot.FileIDs[i] != snapshot2.FileIDs[i] {
			same = false
		}
	}

	return same
}

func ConvertSnapshotArrayToJSON(snapshots []*Snapshot) (string, error) {
	jsonBytes, err := json.Marshal(snapshots)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func ConvertJSONToSnapshotsArray(jsonString string) ([]*Snapshot, error) {
	var snapshots []*Snapshot
	err := json.Unmarshal([]byte(jsonString), &snapshots)
	if err != nil {
		return snapshots, err
	}

	return snapshots, nil
}

func IsSnapshotArraysEqual(snapshots1 []*Snapshot, snapshots2 []*Snapshot) bool {
	if snapshots1 == nil || snapshots2 == nil {
		return false
	}

	if len(snapshots1) != len(snapshots2) {
		return false
	}

	counter := 0
	for i := range snapshots1 {
		if snapshots1[i].Equals(snapshots2[i]) {
			counter++
		}
	}

	if counter == len(snapshots2) {
		return true
	}

	return false
}

func (snapshot *Snapshot) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
