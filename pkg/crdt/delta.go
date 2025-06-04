package crdt

type OperationType string

const (
	OPSetKeyValue    OperationType = "setkeyvalue"
	OPRemoveKeyValue OperationType = "removekeyvalue"
	OPAddEdge        OperationType = "addedge"
	OPRemoveEdge     OperationType = "removeedge"
	OPSetLiteral     OperationType = "setliteral"
	OPMarkDeleted    OperationType = "markdeleted"
	OPCreateNode     OperationType = "createnode"
	// Add more if needed
)

type Delta struct {
	OpType OperationType `json:"op_type"`

	// Common fields
	TargetID NodeID   `json:"target_id,omitempty"` // the node being operated on
	ClientID ClientID `json:"client_id"`
	Version  int      `json:"version"`

	// Fields for SetKeyValue, RemoveKeyValue
	Key   string      `json:"key,omitempty"`
	Value interface{} `json:"value,omitempty"`

	// Fields for AddEdge, RemoveEdge
	From    NodeID `json:"from,omitempty"`
	To      NodeID `json:"to,omitempty"`
	Label   string `json:"label,omitempty"`
	LSEQPos []int  `json:"lseqposition,omitempty"`

	// Fields for CreateNode
	Name     string   `json:"name,omitempty"`
	NodeType NodeType `json:"node_type,omitempty"`
	IsRoot   bool     `json:"isroot,omitempty"`
	ParentID bool     `json:"parentid,omitempty"`
}
