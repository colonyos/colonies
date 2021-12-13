package core

type Colony struct {
	id   string
	name string
}

func CreateColony(id string, name string) (*Colony, error) {
	colony := &Colony{id: id, name: name}

	return colony, nil
}

func (colony *Colony) Name() string {
	return colony.name
}

func (colony *Colony) ID() string {
	return colony.id
}
