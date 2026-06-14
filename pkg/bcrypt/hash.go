package bcrypt

import "golang.org/x/crypto/bcrypt"

func Hash(item string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(item), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckHash(item, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(item))
	return err == nil
}
