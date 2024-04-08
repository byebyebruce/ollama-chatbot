package boltdb

import (
	"encoding/json"
	"os"
	"time"

	"github.com/byebyebruce/ollama-chatbot/pkg/persist"
	"go.etcd.io/bbolt"
)

var _ persist.Persistent = (*BoltDB)(nil)

type BoltDB struct {
	bucketName string
	db         *bbolt.DB
}

func NewBoltDB(path string, name string) (*BoltDB, error) {
	db, err := bbolt.Open(path, os.ModePerm, &bbolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		return err

	})
	if err != nil {
		return nil, err
	}

	return &BoltDB{
		db:         db,
		bucketName: name,
	}, nil
}

func (p *BoltDB) Save(key string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return p.db.Update(func(tx *bbolt.Tx) error {
		bu := tx.Bucket([]byte(p.bucketName))
		return bu.Put([]byte(key), b)
	})
}

func (p *BoltDB) Load(key string, data interface{}) (bool, error) {
	var (
		b []byte
	)
	err := p.db.View(func(tx *bbolt.Tx) error {
		bu := tx.Bucket([]byte(p.bucketName))
		b = bu.Get([]byte(key))
		return nil
	})
	if err != nil {
		return false, err
	}
	if len(b) == 0 {
		return false, nil
	}
	if err = json.Unmarshal(b, data); err != nil {
		return true, err
	}
	return true, nil
}

func (p *BoltDB) Delete(key string) error {
	return p.db.Update(func(tx *bbolt.Tx) error {
		bu := tx.Bucket([]byte(p.bucketName))
		return bu.Delete([]byte(key))
	})
}

func (p *BoltDB) Close() {
	p.db.Close()
}
