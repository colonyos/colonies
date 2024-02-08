package dht

const (
	ADD_CONTACT   int = 0
	FIND_CONTACTS int = 1
	STOP          int = 100
)

type routingTableWorker struct {
	jobQueue chan job
	rt       *routingTable
}

type job struct {
	jobType    int
	contact    Contact
	kademliaID string
	count      int
	res        chan []Contact
}

func createRoutingTableWorker(contact Contact) *routingTableWorker {
	rtw := &routingTableWorker{}
	rtw.jobQueue = make(chan job, 1000)
	rtw.rt = createRoutingTable(contact)

	return rtw
}

func (rtw *routingTableWorker) serveForever() {
	for {
		job := <-rtw.jobQueue
		switch job.jobType {
		case ADD_CONTACT:
			rtw.rt.addContact(job.contact)
		case FIND_CONTACTS:
			job.res <- rtw.rt.findClosestContacts(CreateKademliaID(job.kademliaID), job.count)
		case STOP:
			return
		}
	}
}

func (rtw *routingTableWorker) addContact(contact Contact) {
	rtw.jobQueue <- job{jobType: ADD_CONTACT, contact: contact}
}

func (rtw *routingTableWorker) findContacts(kademliaID string, count int) chan []Contact {
	res := make(chan []Contact)
	rtw.jobQueue <- job{jobType: FIND_CONTACTS, kademliaID: kademliaID, res: res, count: count}
	return res
}

func (rtw *routingTableWorker) shutdown() {
	rtw.jobQueue <- job{jobType: STOP}
}
