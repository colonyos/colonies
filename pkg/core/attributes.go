package core

const (
	IN  int = 0
	OUT     = 1
	ERR     = 2
)

type Attribute struct {
	taskID   string
	taskType int
	key      string
	value    string
}
