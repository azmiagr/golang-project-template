package bcrypt

import lib_bcrypt "golang.org/x/crypto/bcrypt"

type Interface interface {
	GenerateFromPassword(password string) (string, error)
	CompareAndHashPassword(hashPassword, password string) error
}

type bcrypt struct {
	cost int
}

func Init() Interface {
	return &bcrypt{
		cost: 10,
	}
}

func (b *bcrypt) GenerateFromPassword(password string) (string, error) {
	bytePass, err := lib_bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", err
	}

	return string(bytePass), nil
}

func (b *bcrypt) CompareAndHashPassword(hashPassword, password string) error {
	err := lib_bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	if err != nil {
		return err
	}

	return nil
}
