package models

import (
	"log"
	"os"
	"strings"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/sqids/sqids-go"
)

func ReturnAlphabet() (string, error) {
	alphabet := os.Getenv("LEMC_SQUID_ALPHABET")
	if alphabet == "" {
		panic("LEMC_SQUID_ALPHABET is not set")
	}
	return alphabet, nil
}

func SquidAndNameByAccountID(id int64) (string, string, error) {
	alphabet, err := ReturnAlphabet()
	if err != nil {
		return "", "", err
	}

	log.Println("lemon squid alphabet", alphabet)
	account := &Account{}
	query := "SELECT * FROM accounts WHERE id = ?"
	err = db.Db().Get(account, query, id)
	if err != nil {
		log.Println("Error:", err)
		return "", "", err
	}

	s, err := sqids.New(
		sqids.Options{
			Blocklist: nil,
			MinLength: 4,
			Alphabet:  alphabet,
		})

	if err != nil {
		return "", "", err
	}
	uid := []uint64{uint64(account.ID)}
	sid, err := s.Encode(uid)
	log.Println("lemon squid sid", sid)
	if err != nil {
		return "", "", err
	}

	return sid, strings.ToLower(account.Name), nil
}
