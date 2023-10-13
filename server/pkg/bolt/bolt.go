package bolt

import (
	"time"

	bolt "go.etcd.io/bbolt"
)

func NewBolt(path string) (*bolt.DB, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	return db, nil
}
