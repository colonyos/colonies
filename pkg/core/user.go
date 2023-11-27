package core

import (
	"encoding/json"
)

type User struct {
	ColonyName string `json:"colonyname"`
	ID         string `json:"userid"`
	Name       string `json:"name"`
}

func CreateUser(colonyName string, userID string, name string) *User {
	return &User{ColonyName: colonyName, ID: userID, Name: name}
}

func ConvertJSONToUser(jsonString string) (*User, error) {
	var user *User
	err := json.Unmarshal([]byte(jsonString), &user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func ConvertJSONToUserArray(jsonString string) ([]*User, error) {
	var users []*User

	err := json.Unmarshal([]byte(jsonString), &users)
	if err != nil {
		return users, err
	}

	return users, nil
}

func ConvertUserArrayToJSON(users []*User) (string, error) {
	jsonBytes, err := json.Marshal(users)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func IsUserArraysEqual(users1 []*User, users2 []*User) bool {
	counter := 0
	for _, user1 := range users1 {
		for _, user2 := range users2 {
			if user1.Equals(user2) {
				counter++
			}
		}
	}

	if counter == len(users1) && counter == len(users2) {
		return true
	}

	return false
}

func (user *User) Equals(user2 *User) bool {
	if user2 == nil {
		return false
	}

	if user.ID == user2.ID &&
		user.ColonyName == user2.ColonyName &&
		user.Name == user2.Name {
		return true
	}

	return false
}

func (user *User) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(user)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
