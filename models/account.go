package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jmoiron/sqlx"
	"github.com/sqids/sqids-go"
)

type Account struct {
	Created   time.Time `db:"created"`
	Updated   time.Time `db:"updated"`
	ID        int64     `db:"id"`
	Squid     string    `db:"squid"`
	Name      string    `db:"name"`
	IsDeleted bool      `db:"is_deleted"`
}

func AccountBySquid(squid string) (*Account, error) {
	account := &Account{}
	s, err := sqids.New(
		sqids.Options{
			Blocklist: nil,
			MinLength: 4,
			Alphabet:  os.Getenv("LEMC_SQUID_ALPHABET"),
		})

	if err != nil {
		return nil, fmt.Errorf("could not initialize squid generator: %w", err)
	}

	id := s.Decode(squid)
	if len(id) != 1 {
		return nil, fmt.Errorf("invalid squid format")
	}

	query := "SELECT id, name, squid, created, updated, is_deleted FROM accounts WHERE id = ?"
	err = db.Db().Get(account, query, id[0])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err // Return sql.ErrNoRows specifically if needed upstream
		}
		return nil, fmt.Errorf("database error fetching account by squid: %w", err)
	}

	return account, nil
}

func AccountByID(id int64) (*Account, error) {
	account := &Account{}
	query := "SELECT id, name, squid, created, updated, is_deleted FROM accounts WHERE id = ?"
	err := db.Db().Get(account, query, id)
	if err != nil {
		log.Println("Error:", err)
		return account, err
	}

	return account, nil
}

func AccountCreate(name string, tx *sqlx.Tx) (*Account, error) {
	s, err := sqids.New(sqids.Options{
		MinLength: 4,
		Alphabet:  os.Getenv("LEMC_SQUID_ALPHABET"),
	})
	if err != nil {
		log.Println("Error creating sqids generator:", err)
		return nil, err // Return error as squid generation failed
	}

	// Get the next ID that will be created
	var id int64
	err = tx.Get(&id, "SELECT seq + 1 FROM sqlite_sequence WHERE name = 'accounts'")
	if err != nil {
		if err == sql.ErrNoRows {
			id = 1
		} else {
			log.Println("Error getting next ID:", err)
			return nil, err
		}
	}

	squid, err := s.Encode([]uint64{uint64(id)})
	if err != nil {
		log.Println("Error encoding squid:", err)
		return nil, err // Return error as squid encoding failed
	}

	account := &Account{
		Name:  name,
		Squid: squid,
	}

	query := "INSERT INTO accounts (name, squid) VALUES (?, ?)"
	res, err := tx.Exec(query, name, squid)
	if err != nil {
		return account, err
	}

	lid, err := res.LastInsertId()
	if err != nil {
		return account, err
	}

	if lid != id {
		return account, fmt.Errorf("last insert id does not match expected id: %d != %d", lid, id)
	}

	account.ID = lid
	account.Squid = squid

	return account, nil
}
