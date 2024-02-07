package dht

const (
	ADD_CONTACT   int = 0
	FIND_CONTACTS int = 1
)

type routingTableWorker struct {
	jobQueue     chan *job
	routingTable *routingTable
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
	rtw.jobQueue = make(chan *job, 1000)
	rtw.routingTable = createRoutingTable(contact)
	go rtw.serveForever()

	return rtw
}

func (rtw *routingTableWorker) serveForever() {
	for {
		job := <-rtw.jobQueue
		switch job.jobType {
		case ADD_CONTACT:
			rtw.routingTable.addContact(job.contact)
		case FIND_CONTACTS:
			job.res <- rtw.routingTable.findClosestContacts(CreateKademliaID(job.kademliaID), job.count) // RACE CONDITION
		}
	}
}

func (rtw *routingTableWorker) addContact(contact Contact) {
	rtw.jobQueue <- &job{jobType: ADD_CONTACT, contact: contact}
}

func (rtw *routingTableWorker) findContacts(kademliaID string, count int) chan []Contact {
	res := make(chan []Contact)
	rtw.jobQueue <- &job{jobType: FIND_CONTACTS, kademliaID: kademliaID, res: res, count: count}
	return res
}
