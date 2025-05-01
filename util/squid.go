package util

import (
	"log"
	"os"
	"strings"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/sqids/sqids-go"
)

func SquidAndNameByAccountID(id int64) (string, string, error) {
	account := &models.Account{}
	query := "SELECT * FROM accounts WHERE id = ?"
	err := db.Db().Get(account, query, id)
	if err != nil {
		log.Println("Error:", err)
		return "", "", err
	}

	s, err := sqids.New(
		sqids.Options{
			Blocklist: nil,
			MinLength: 4,
			Alphabet:  os.Getenv("LEMC_SQUID_ALPHABET"),
		})

	if err != nil {
		return "", "", err
	}

	sid, err := s.Encode([]uint64{uint64(account.ID)})
	if err != nil {
		return "", "", err
	}

	return sid, strings.ToLower(account.Name), nil
}
