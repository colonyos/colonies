package core

type ProcessGraphSpec struct {
	Group       bool          `json:"group"`
	ProcessSpec []ProcessSpec `json:"processspecs"`
}
