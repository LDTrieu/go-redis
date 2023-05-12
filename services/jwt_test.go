package services

import (
	"log"
	"os"
	"testing"
)

func TestNewJWT(t *testing.T) {
	key, _ := os.ReadFile("private_key.pem")
	log.Println(key)
}
