package utils

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
	ErrNotMatch            = errors.New("notmatch of argon2")
)

var (
	Pw passworder
)

func init() {
	Pw = &bcryptImpl{
		cost: bcrypt.DefaultCost,
	}
}

type passworder interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword string,
		candidatePassword string) error
}

type bcryptImpl struct {
	cost int
}

func (b *bcryptImpl) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)

	if err != nil {
		return "", fmt.Errorf("could not hash password %w", err)
	}
	return string(hashedPassword), nil
}

func (b *bcryptImpl) VerifyPassword(hashedPassword string, candidatePassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(candidatePassword))
}
