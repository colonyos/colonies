package crdt

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

func lowestClientID(a, b ClientID) ClientID {
	if a < b {
		return a
	}
	return b
}

func normalizeNumber(v interface{}) interface{} {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return v
	}
}

func setNodeTypeFlags(node *NodeCRDT, nodeType NodeType) {
	switch nodeType {
	case Root:
		node.IsRoot = true
	case Map:
		node.IsMap = true
	case Array:
		node.IsArray = true
	case Literal:
		node.IsLiteral = true
	default:
		log.WithField("NodeType", nodeType).Error("Unknown node type, defaulting to literal")
		node.IsLiteral = true
	}
}

func buildOpString(opName string, args ...interface{}) string {
	if len(args) == 0 {
		return opName + "()"
	}

	str := opName + "("
	for i, arg := range args {
		if i > 0 {
			str += ", "
		}
		str += fmt.Sprintf("%v", arg)
	}
	str += ")"
	return str
}
