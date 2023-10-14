package db

import (
	"server/internal/db/bbolt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	db, _ := bbolt.NewBoltDb("test")
	defer db.Close()

	repo := NewRepository(db)

	repo.Save("test", 100)

	ids := repo.Get()

	assert.Equal(t, "test", ids[0])

	repo.Save("test3", 100)
	repo.Save("test4", 100)

	ids = repo.Get()

	assert.Equal(t, 3, len(ids))
}

func TestSave(t *testing.T) {
	db, _ := bbolt.NewBoltDb("test")
	defer db.Close()

	repo := NewRepository(db)

	err := repo.Save("test", 100)

	assert.Equal(t, nil, err)
}
