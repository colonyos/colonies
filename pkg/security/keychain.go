package security

import (
	"errors"
	"io/ioutil"
	"os"
)

type Keychain struct {
	dirName string
}

func (keychain *Keychain) ensureColoniesDirExists() error {
	err := os.Mkdir(keychain.dirName, 0700)
	if err == nil {
		return nil
	}
	if os.IsExist(err) {
		info, err := os.Stat(keychain.dirName)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return errors.New(keychain.dirName + " exists but is not a directory")
		}
		return nil
	}
	return err
}

func CreateKeychain(coloniesDirName string) (*Keychain, error) {
	keychain := &Keychain{}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	keychain.dirName = homeDir + "/" + coloniesDirName

	err = keychain.ensureColoniesDirExists()
	if err != nil {
		return nil, err
	}

	return keychain, nil
}

func (keychain *Keychain) AddPrvKey(id string, prvKey string) error {
	prvKeyBytes := []byte(prvKey)
	return os.WriteFile(keychain.dirName+"/"+id, prvKeyBytes, 0600)
}

func (keychain *Keychain) GetPrvKey(id string) (string, error) {
	prvKeyBytes, err := ioutil.ReadFile(keychain.dirName + "/" + id)
	if err != nil {
		return "", err
	}

	return string(prvKeyBytes), nil
}

func (keychain *Keychain) Remove() error {
	return os.RemoveAll(keychain.dirName)
}
