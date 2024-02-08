package dht

import (
	"container/list"
)

type bucket struct {
	list *list.List
}

func createBucket() *bucket {
	bucket := &bucket{}
	bucket.list = list.New()
	return bucket
}

func (bucket *bucket) addContact(contact Contact) {
	var element *list.Element
	for e := bucket.list.Front(); e != nil; e = e.Next() {
		nodeID := e.Value.(Contact).ID

		if (contact).ID.Equals(nodeID) {
			element = e
		}
	}

	if element == nil {
		if bucket.list.Len() < bucketSize {
			bucket.list.PushFront(contact)
		} else {
			// TODO: Check if the last contact is alive
		}
	} else {
		bucket.list.MoveToFront(element)
	}
}

func (bucket *bucket) removeContact(contact Contact) {
	for e := bucket.list.Front(); e != nil; e = e.Next() {
		nodeID := e.Value.(Contact).ID

		if (contact).ID.Equals(nodeID) {
			bucket.list.Remove(e)
		}
	}
}

func (bucket *bucket) getContact(target KademliaID) Contact {
	for elt := bucket.list.Front(); elt != nil; elt = elt.Next() {
		contact := elt.Value.(Contact)
		if contact.ID.Equals(target) {
			return contact
		}
	}
	return Contact{}
}

func (bucket *bucket) getContactAndCalcDistance(target KademliaID) []Contact {
	var contacts []Contact

	for elt := bucket.list.Front(); elt != nil; elt = elt.Next() {
		contact := elt.Value.(Contact)
		contact.CalcDistance(target)
		contacts = append(contacts, contact)
	}

	return contacts
}

func (bucket *bucket) len() int {
	return bucket.list.Len()
}
