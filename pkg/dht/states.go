package dht

const (
	ADD_CONTACT   int = 0
	FIND_CONTACTS int = 1
	PUT           int = 2
	GET           int = 3
	STOP          int = 100
)

type states struct {
	jobQueue chan job
	rt       *routingTable
	kvs      *kvStore
}

type job struct {
	jobType      int
	contact      Contact
	kademliaID   string
	count        int
	contactsChan chan []Contact
	id           string
	key          string
	value        string
	sig          string
	kvChan       chan []KV
	errChan      chan error
}

func createStates(contact Contact) *states {
	s := &states{}
	s.jobQueue = make(chan job, 1000)
	s.rt = createRoutingTable(contact)
	s.kvs = createKVStore()
	return s
}

func (s *states) serveForever() {
	for {
		job := <-s.jobQueue
		switch job.jobType {
		case ADD_CONTACT:
			s.rt.addContact(job.contact)
			job.errChan <- nil
		case FIND_CONTACTS:
			job.contactsChan <- s.rt.findClosestContacts(CreateKademliaID(job.kademliaID), job.count)
		case PUT:
			job.errChan <- s.kvs.put(job.id, job.key, job.value, job.sig)
		case GET:
			kvs, err := s.kvs.getAllValuesWithPrefix(job.value)
			if err != nil {
				job.errChan <- err
			} else {
				job.kvChan <- kvs
			}
		case STOP:
			return
		}
	}
}

func (s *states) addContact(contact Contact) chan error {
	errChan := make(chan error, 1)
	s.jobQueue <- job{jobType: ADD_CONTACT, contact: contact, errChan: errChan}
	return errChan
}

func (s *states) findContacts(kademliaID string, count int) (chan []Contact, chan error) {
	contactsChan := make(chan []Contact, 1)
	errChan := make(chan error, 1)
	s.jobQueue <- job{jobType: FIND_CONTACTS, kademliaID: kademliaID, contactsChan: contactsChan, errChan: errChan, count: count}
	return contactsChan, errChan
}

func (s *states) put(id string, key string, value string, sig string) chan error {
	errChan := make(chan error, 1)
	s.jobQueue <- job{jobType: PUT, id: id, key: key, value: value, sig: sig, errChan: errChan}
	return errChan
}

func (s *states) get(value string) (chan []KV, chan error) {
	kvChan := make(chan []KV, 1)
	errChan := make(chan error, 1)
	s.jobQueue <- job{jobType: GET, value: value, errChan: errChan, kvChan: kvChan}
	return kvChan, errChan
}

func (s *states) shutdown() {
	s.jobQueue <- job{jobType: STOP}
}
