package gkv

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
)

type KeyValueStore interface {
	Get(key string) (string, error)
	Set(key string, value string) error

	GetUInt64(key string) (uint64, error)
	SetUInt64(key string, value uint64) error

	Del(key string) error

	Close() error
}

type SQLKeyValueStoreConfig struct {
	// database table details
	TableName   string `yaml:"table_name"`
	KeyColumn   string `yaml:"key_column"`
	ValueColumn string `yaml:"value_column"`

	// limits
	MaxKeyLen   int `yaml:"max_key_len"`
	MaxValueLen int `yaml:"max_value_len"`
}

type SQLKeyValueStore struct {
	config      *SQLKeyValueStoreConfig
	maxKeyLen   int
	maxValueLen int

	stmtGet *sql.Stmt
	stmtSet *sql.Stmt
	stmtDel *sql.Stmt
}

type closable interface {
	Close() error
}

func NewMySQLKeyValueStore(connection *sql.DB, config *SQLKeyValueStoreConfig) (*SQLKeyValueStore, error) {
	var err error

	keyValueStore := &SQLKeyValueStore{
		config:      config,
		maxKeyLen:   config.MaxKeyLen,
		maxValueLen: config.MaxValueLen,
	}

	keyValueStore.stmtGet, err = connection.Prepare(fmt.Sprintf("SELECT `%s` FROM `%s` WHERE `%s` = BINARY ?", config.ValueColumn, config.TableName, config.KeyColumn))
	if err != nil {
		return nil, err
	}

	keyValueStore.stmtSet, err = connection.Prepare(fmt.Sprintf("INSERT INTO `%s`(`%s`, `%s`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `%s` = VALUES(`%s`)", config.TableName, config.KeyColumn, config.ValueColumn, config.ValueColumn, config.ValueColumn))
	if err != nil {
		return nil, err
	}

	keyValueStore.stmtDel, err = connection.Prepare(fmt.Sprintf("DELETE FROM `%s` WHERE `%s` = ?", config.TableName, config.KeyColumn))
	if err != nil {
		return nil, err
	}

	return keyValueStore, nil
}

func (s *SQLKeyValueStore) Get(key string) (string, error) {
	var value string

	row := s.stmtGet.QueryRow(key)
	err := row.Scan(&value)

	if err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return value, nil
}

func (s *SQLKeyValueStore) Set(key string, value string) error {
	if len(key) > s.maxKeyLen {
		return fmt.Errorf("Key is too long")
	}
	if len(value) > s.maxValueLen {
		return fmt.Errorf("Value is too long")
	}
	_, err := s.stmtSet.Exec(key, value)

	return err
}

func (s *SQLKeyValueStore) GetUInt64(key string) (uint64, error) {
	data := make([]byte, 8)

	row := s.stmtGet.QueryRow(key)
	err := row.Scan(&data)

	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	var num uint64
	err = binary.Read(bytes.NewBuffer(data[:]), binary.LittleEndian, &num)
	if err != nil {
		return 0, err
	}

	return num, nil
}

func (s *SQLKeyValueStore) SetUInt64(key string, value uint64) error {
	data := make([]byte, 8)

	binary.LittleEndian.PutUint64(data, value)
	_, err := s.stmtSet.Exec(key, data)

	return err
}

func (s *SQLKeyValueStore) Del(key string) error {
	_, err := s.stmtDel.Exec(key)
	return err
}

func (s *SQLKeyValueStore) Close() error {
	for _, c := range []*sql.Stmt{s.stmtSet, s.stmtGet, s.stmtDel} {
		if err := c.Close(); err != nil {
			return err
		}
	}

	return nil
}
