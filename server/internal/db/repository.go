package db

import (
	"errors"
	"log"
	"strconv"

	bolt "go.etcd.io/bbolt"
)

type Repository interface {
	Get() []string
	Save(id string, period uint64) error
}

type repository struct {
	bolt *bolt.DB
}

func NewRepository(db *bolt.DB) Repository {
	return &repository{
		bolt: db,
	}
}

func (r *repository) Get() []string {
	path := []byte(r.bolt.Path())
	ids := make([]string, 0, 100)
	r.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(path)
		if b == nil {
			return errors.New("bucket doesn't exist")
		}
		err := b.ForEach(func(k, v []byte) error {
			ids = append(ids, string(k))
			return nil
		})
		return err
	})
	return ids
}

func (r *repository) Save(id string, period uint64) error {
	path := []byte(r.bolt.Path())
	return r.bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(path)
		if err != nil {
			log.Println("tx.CreateBucketIfNotExists err ", err)
			return err
		}
		value := strconv.FormatInt(int64(period), 10)
		b.Put([]byte(id), []byte(value))
		return nil
	})
}
