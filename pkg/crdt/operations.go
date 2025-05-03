package crdt

type VectorClockEntry struct {
	ClientID string `json:"clientid"`
	Version  int    `json:"version"`
}

type AddNode struct {
	Name        string           `json:"name"`
	IsArray     bool             `json:"isarray"`
	ParentID    string           `json:"parentid"`
	VectorClock VectorClockEntry `json:"vectorclock"`
}

type AddEdge struct {
	FromNodeID  string           `json:"fromnodeid"`
	ToNodeID    string           `json:"tonodeid"`
	Label       string           `json:"label"`
	VectorClock VectorClockEntry `json:"vectorclock"`
}

type InsertEdge struct {
	FromNodeID  string           `json:"fromnodeid"`
	ToNodeID    string           `json:"tonodeid"`
	Label       string           `json:"label"`
	Position    int              `json:"position"`
	VectorClock VectorClockEntry `json:"vectorclock"`
}

type RemoveEdge struct {
	FromNodeID  string           `json:"fromnodeid"`
	ToNodeID    string           `json:"tonodeid"`
	VectorClock VectorClockEntry `json:"vectorclock"`
}

type SetField struct {
	Key         string           `json:"key"`
	Value       interface{}      `json:"value"`
	VectorClock VectorClockEntry `json:"vectorclock"`
}

type SetLiteral struct {
	Value       interface{}      `json:"value"`
	VectorClock VectorClockEntry `json:"vectorclock"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
}
