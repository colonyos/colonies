package dht

import (
	"testing"
)

func TestBucketAddContact(t *testing.T) {
	bucket := createBucket()
	contact1 := Contact{ID: NewRandomKademliaID()}
	contact2 := Contact{ID: NewRandomKademliaID()}

	bucket.addContact(contact1)
	if bucket.len() != 1 {
		t.Errorf("Expected bucket length 1, got %d", bucket.len())
	}

	bucket.addContact(contact2)
	if bucket.len() != 2 {
		t.Errorf("Expected bucket length 2, got %d", bucket.len())
	}

	// Test adding a duplicate contact
	bucket.addContact(contact1)
	if bucket.len() != 2 {
		t.Errorf("Expected bucket length 2 after adding a duplicate, got %d", bucket.len())
	}
}

func TestBucketRemoveContact(t *testing.T) {
	bucket := createBucket()
	contact1 := Contact{ID: NewRandomKademliaID()}
	contact2 := Contact{ID: NewRandomKademliaID()}

	bucket.addContact(contact1)
	bucket.addContact(contact2)
	bucket.removeContact(contact1)

	if bucket.len() != 1 {
		t.Errorf("Expected bucket length 1 after removal, got %d", bucket.len())
	}

	bucket.removeContact(contact1)
	if bucket.len() != 1 {
		t.Errorf("Expected bucket length 1 after removing a non-existing contact, got %d", bucket.len())
	}
}

func TestBucketGetContact(t *testing.T) {
	bucket := createBucket()
	kademliaID1 := NewRandomKademliaID()
	kademliaID2 := NewRandomKademliaID()
	contact1 := Contact{ID: kademliaID1}
	bucket.addContact(contact1)

	result := bucket.getContact(kademliaID1)
	if result.ID != contact1.ID {
		t.Errorf("Expected to get contact with ID %s, got %s", contact1.ID, result.ID)
	}

	// Test getting a non-existing contact
	result = bucket.getContact(kademliaID2)
	if result.ID.String() == "" {
		t.Errorf("Expected to get an empty contact for a non-existing ID, got %s", result.ID)
	}
}

func TestBucketGetContactAndCalcDistance(t *testing.T) {
	// TODO
}
