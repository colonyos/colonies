package dht

const bucketSize = 20

type routingTable struct {
	me      Contact
	buckets [IDLength * 8]*bucket
}

func createRoutingTable(me Contact) *routingTable {
	routingTable := &routingTable{}
	for i := 0; i < IDLength*8; i++ {
		routingTable.buckets[i] = createBucket()
	}
	routingTable.me = me
	return routingTable
}

func (routingTable *routingTable) addContact(contact Contact) {
	bucketIndex := routingTable.getBucketIndex(contact.ID)
	bucket := routingTable.buckets[bucketIndex]
	bucket.addContact(contact)
}

func (routingTable *routingTable) removeContact(contact Contact) {
	bucketIndex := routingTable.getBucketIndex(contact.ID)
	bucket := routingTable.buckets[bucketIndex]
	bucket.removeContact(contact)
}

func (routingTable *routingTable) getContact(id KademliaID) Contact {
	bucketIndex := routingTable.getBucketIndex(id)
	bucket := routingTable.buckets[bucketIndex]
	return bucket.getContact(id)
}

func (routingTable *routingTable) findClosestContacts(target KademliaID, count int) []Contact {
	var candidates ContactCandidates
	bucketIndex := routingTable.getBucketIndex(target)
	bucket := routingTable.buckets[bucketIndex]

	candidates.Append(bucket.getContactAndCalcDistance(target))

	for i := 1; (bucketIndex-i >= 0 || bucketIndex+i < IDLength*8) && candidates.Len() < count; i++ {
		if bucketIndex-i >= 0 {
			bucket = routingTable.buckets[bucketIndex-i]
			candidates.Append(bucket.getContactAndCalcDistance(target))
		}
		if bucketIndex+i < IDLength*8 {
			bucket = routingTable.buckets[bucketIndex+i]
			candidates.Append(bucket.getContactAndCalcDistance(target))
		}
	}

	candidates.Sort()

	if count > candidates.Len() {
		count = candidates.Len()
	}

	return candidates.GetContacts(count)
}

func (routingTable *routingTable) getBucketIndex(id KademliaID) int {
	distance := id.CalcDistance(routingTable.me.ID)
	for i := 0; i < IDLength; i++ {
		for j := 0; j < 8; j++ {
			if (distance[i]>>uint8(7-j))&0x1 != 0 {
				return i*8 + j
			}
		}
	}

	return IDLength*8 - 1
}
