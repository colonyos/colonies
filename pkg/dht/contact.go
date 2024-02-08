package dht

import (
	"fmt"
	"sort"
)

type Contact struct {
	ID       KademliaID `json:"kademliaid"`
	Addr     string     `json:"address"`
	distance KademliaID
}

func CreateContact(id KademliaID, address string) Contact {
	return Contact{id, address, KademliaID{}}
}

func (contact *Contact) CalcDistance(target KademliaID) {
	contact.distance = contact.ID.CalcDistance(target)
}

func (contact *Contact) Less(otherContact Contact) bool {
	return contact.distance.Less(otherContact.distance)
}

func (contact *Contact) String() string {
	return fmt.Sprintf(`contact("%s", "%s")`, contact.ID, contact.Addr)
}

type ContactCandidates struct {
	contacts []Contact
}

func (candidates *ContactCandidates) Append(contacts []Contact) {
	candidates.contacts = append(candidates.contacts, contacts...)
}

func (candidates *ContactCandidates) GetContacts(count int) []Contact {
	return candidates.contacts[:count]
}

func (candidates *ContactCandidates) Sort() {
	sort.Sort(candidates)
}

func (candidates *ContactCandidates) Len() int {
	return len(candidates.contacts)
}

func (candidates *ContactCandidates) Swap(i, j int) {
	candidates.contacts[i], candidates.contacts[j] = candidates.contacts[j], candidates.contacts[i]
}

func (candidates *ContactCandidates) Less(i, j int) bool {
	return candidates.contacts[i].Less(candidates.contacts[j])
}
