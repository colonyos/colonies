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
	key          string
	value        string
	valuesChan   chan []string
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
		case FIND_CONTACTS:
			job.contactsChan <- s.rt.findClosestContacts(CreateKademliaID(job.kademliaID), job.count)
		case PUT:
			job.errChan <- s.kvs.put(job.key, job.value)
		case GET:
			values, err := s.kvs.getAllValuesWithPrefix(job.value)
			if err != nil {
				job.errChan <- err
			} else {
				job.valuesChan <- values
			}
		case STOP:
			return
		}
	}
}

func (s *states) addContact(contact Contact) {
	s.jobQueue <- job{jobType: ADD_CONTACT, contact: contact}
}

func (s *states) findContacts(kademliaID string, count int) chan []Contact {
	contactsChan := make(chan []Contact)
	s.jobQueue <- job{jobType: FIND_CONTACTS, kademliaID: kademliaID, contactsChan: contactsChan, count: count}
	return contactsChan
}

func (s *states) put(key string, value string) chan error {
	errChan := make(chan error)
	s.jobQueue <- job{jobType: PUT, key: key, value: value, errChan: errChan}
	return errChan
}

func (s *states) get(value string) (chan []string, chan error) {
	valuesChan := make(chan []string)
	errChan := make(chan error)
	s.jobQueue <- job{jobType: GET, value: value, errChan: errChan, valuesChan: valuesChan}
	return valuesChan, errChan
}

func (s *states) shutdown() {
	s.jobQueue <- job{jobType: STOP}
}
