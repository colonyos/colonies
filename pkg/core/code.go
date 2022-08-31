package core

type Code struct {
	ID     string `json:"codeid"`
	Name   string `json:"name"`
	Bundle []byte `json:"bundle"`
	Script string `json:"script"`
}

// func CreateCode(name string, package []byte) Code {
//     code := &Code{Name:name, Package:package}
//     return code
// }
