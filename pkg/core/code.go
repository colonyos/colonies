package core

type Code struct {
	ID      string `json:"codeid"`
	Name    string `json:"name"`
	Package []byte `json:"package"`
	 []byte `json:"package"`
}

func CreateCode(name string, package []byte) Code {
    code := &Code{Name:name, Package:package}
    return code
}
