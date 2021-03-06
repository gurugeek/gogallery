package datastore

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/dgraph-io/badger"
	"time"
)

type Album struct {
	Id         string    `json:"id"`
	Name       string    `json:"name"`
	ModTime    time.Time `json:"mod_time"`
	Parent     string    `json:"parent"`
	ProfileIMG *Picture  `json:"profile_image"`
	Images     []Picture `json:"images"`
}

type Directory struct {
	Album    Album        `json:"album"`
	Children []*Directory `json:"children"`
}

func (u *Album) serialize() []byte {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	if err := e.Encode(u); err != nil {
		panic(err)
	}
	return b.Bytes()
}

func (u *Album) deserialize(b []byte) error {
	dCache := bytes.NewBuffer(b)
	d := gob.NewDecoder(dCache)
	if err := d.Decode(u); err != nil {
		return err
	}
	return nil
}

type albumDataStore struct {
	db     *badger.DB
	Albums []Album
}

func (uDs *albumDataStore) Close() {
	uDs.db.Close()
}
func (u albumDataStore) New() *albumDataStore {
	u = albumDataStore{}
	u.Initialize()
	return &u
}

func (u *albumDataStore) Initialize() {
	u.db = createDatastore("album")
}

func (uDs *albumDataStore) Save(u interface{}) error {
	original, ok := u.(Album)
	if ok {
		err := uDs.db.Update(func(tx *badger.Txn) error {
			//fmt.Println(original.Id)
			return tx.Set([]byte(original.Id), original.serialize())
		})
		//fmt.Println(err)
		return err
	}
	fmt.Println("NO ERROR")
	return nil
}

func (uDs *albumDataStore) Edit(u interface{}) error {
	original, ok := u.(Album)
	if ok {
		err := uDs.db.Update(func(tx *badger.Txn) error {
			//fmt.Println(original.Id)
			return tx.Set([]byte(original.Id), original.serialize())
		})
		//fmt.Println(err)
		return err
	}
	fmt.Println("NO ERROR")
	return nil
}

func (uDs *albumDataStore) Delete(u interface{}) error {
	original, ok := u.(Album)
	if ok {
		err := uDs.db.Update(func(tx *badger.Txn) error {
			return tx.Delete([]byte(original.Id))
		})
		return err
	}
	return nil
}

func (uDs *albumDataStore) Get(id string) (interface{}, error) {
	var valCopy []byte
	err := uDs.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get([]byte(id))
		if err != nil {
			return err
		}
		valCopy, err = item.ValueCopy(nil)
		return nil
	})
	if err != nil {
		return Album{}, err
	}
	u := Album{}
	u.deserialize(valCopy)
	return u, nil
}

func (uDs *albumDataStore) GetAll() (interface{}, error) {
	albums := []Album{}
	err := uDs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			data, err := item.Value()
			if err != nil {
				return err
			}
			u := Album{}
			error := u.deserialize(data)
			if error != nil {
				return error
			}
			albums = append(albums, u)
		}
		return nil
	})
	return albums, err
}

func (uDs *albumDataStore) Query(field string, val interface{}, limit int) (interface{}, error) {
	albums := []Album{}
	err := uDs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			if (len(albums) >= limit) && (limit != 0) {
				return nil
			}
			item := it.Item()
			data, err := item.Value()
			if err != nil {
				return err
			}
			u := Album{}
			error := u.deserialize(data)
			if error != nil {
				return error
			}
			if getFieldByName(u, field) == val {
				albums = append(albums, u)
			}
		}
		return nil
	})
	return albums, err
}

func BuildTree(albums []Album) (root *Directory) {
	var nodeTable = map[string]*Directory{}
	for _, album := range albums {
		nodeTable[album.Name] = &Directory{album, []*Directory{}}
	}
	for _, node := range nodeTable {
		parent, exists := nodeTable[node.Album.Parent]
		if !exists { // If a parent does not exist, this is the root.
			root = node
		} else {
			parent.Children = append(parent.Children, node)
			//Cache.Tables("ALBUM").Save(Album{path, node.Info.Name, node.Info.ModTime, parent.Info.Name})

		}
	}
	return
}
