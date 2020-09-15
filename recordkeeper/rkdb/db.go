package rkdb

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/zinic/forculus/errors"

	"github.com/dgraph-io/badger/v2"
)

const (
	ErrEventNotFound = errors.New("event not found")

	eventRecordIDKey = "events.next_id"
	eventRecordKey   = "events.id_%d"
)

func formatRecordKey(id uint64) []byte {
	return []byte(fmt.Sprintf(eventRecordKey, id))
}

func nextEventID(txn *badger.Txn) (uint64, error) {
	var (
		value            = make([]byte, 8)
		currentID uint64 = 0
	)

	if item, err := txn.Get([]byte(eventRecordIDKey)); err != nil {
		if err != badger.ErrKeyNotFound {
			return 0, err
		}
	} else if _, err := item.ValueCopy(value); err != nil {
		return 0, err
	} else {
		currentID = binary.LittleEndian.Uint64(value)
	}

	currentID += 1
	binary.LittleEndian.PutUint64(value, currentID)

	return currentID, txn.Set([]byte(eventRecordIDKey), value)
}

func NewDatabase(path string) (*Database, error) {
	if bdb, err := badger.Open(badger.DefaultOptions(path)); err != nil {
		return nil, err
	} else {
		return &Database{
			db: bdb,
		}, nil
	}
}

type Database struct {
	db *badger.DB
}

func (s *Database) Close() error {
	return s.db.Close()
}

func (s *Database) WriteEventRecord(record EventRecord) (uint64, error) {
	txn := s.db.NewTransaction(true)
	defer txn.Discard()

	if eventID, err := nextEventID(txn); err != nil {
		return 0, err
	} else {
		record.ID = eventID

		if output, err := json.Marshal(&record); err != nil {
			return 0, err
		} else if err := txn.Set(formatRecordKey(record.ID), output); err != nil {
			return 0, err
		}

		return eventID, txn.Commit()
	}
}

func (s *Database) GetEventRecord(id uint64) (EventRecord, error) {
	var (
		txn    = s.db.NewTransaction(false)
		record EventRecord
	)

	defer txn.Discard()

	if item, err := txn.Get(formatRecordKey(id)); err != nil {
		if err == badger.ErrKeyNotFound {
			return record, ErrEventNotFound
		}

		return record, err
	} else if value, err := item.ValueCopy(nil); err != nil {
		return record, err
	} else if err := json.Unmarshal(value, &record); err != nil {
		return record, err
	}

	return record, nil
}
