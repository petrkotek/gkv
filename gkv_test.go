package gkv

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"

	// import mysql driver
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func TestMain(m *testing.M) {
	var err error

	// setUp
	db, err = setUpMySQL()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	result := m.Run()

	// tearDown
	err = tearDownMySQL()
	if err != nil {
		fmt.Print(err)
	}

	os.Exit(result)
}

func TestSQLKeyValueStore_SetDoesntReturnAnError(t *testing.T) {
	kv := setUp(t)
	defer tearDown(db)
	defer kv.Close()

	err := kv.Set("foo", "bar")
	if err != nil {
		t.Fatal("Set(foo, bar) shouldn't return an error, got", err)
	}
}

func TestSQLKeyValueStore_GetAfterSetReturnsTheValue(t *testing.T) {
	kv := setUp(t)
	defer tearDown(db)
	defer kv.Close()

	_ = kv.Set("foo", "bar")

	value, err := kv.Get("foo")

	if err != nil {
		t.Fatal("Get(foo) shouldn't return an error, got", err)
	}
	if value != "bar" {
		t.Fatal("Get(foo) should return 'bar', got", value)
	}
}

func TestSQLKeyValueStore_GetNonExistingReturnsEmptyString(t *testing.T) {
	kv := setUp(t)
	defer tearDown(db)
	defer kv.Close()

	value, err := kv.Get("non-existing")
	if err != nil {
		t.Fatal("Get(non-existing) shouldn't return an error, got", err)
	}
	if value != "" {
		t.Fatal("Get(non-existing) should return empty string, got", value)
	}
}

func TestSQLKeyValueStore_DelNonExistingDoesntReturnAnError(t *testing.T) {
	kv := setUp(t)
	defer tearDown(db)
	defer kv.Close()

	err := kv.Del("non-existing")
	if err != nil {
		t.Fatal("Get(non-existing) shouldn't return an error, got", err)
	}
}

func TestSQLKeyValueStore_DelDeletesTheValue(t *testing.T) {
	kv := setUp(t)
	defer tearDown(db)
	defer kv.Close()
	_ = kv.Set("foo", "bar")

	err := kv.Del("foo")
	if err != nil {
		t.Fatal("Get(foo) shouldn't return an error, got", err)
	}

	value, err := kv.Get("foo")
	if err != nil {
		t.Fatal("Get(foo) shouldn't return an error, got", err)
	}
	if value != "" {
		t.Fatal("Get(foo) should return empty string, got", value)
	}
}

func TestSQLKeyValueStore_SecondSetOverwritesOriginalValue(t *testing.T) {
	kv := setUp(t)
	defer tearDown(db)
	defer kv.Close()
	_ = kv.Set("foo", "original")

	err := kv.Set("foo", "new")
	if err != nil {
		t.Fatal("Set(foo, new) shouldn't return an error, got", err)
	}

	value, err := kv.Get("foo")
	if err != nil {
		t.Fatal("Get(foo) shouldn't return an error, got", err)
	}
	if value != "new" {
		t.Fatal("Get(foo) should return 'new', got", value)
	}
}

func TestSQLKeyValueStore_UsingLongKeyFails(t *testing.T) {
	kv := setUp(t)
	defer tearDown(db)
	defer kv.Close()

	// we limit the key length 8 characters in tests
	err := kv.Set("123456789", "bar")
	if err == nil {
		t.Error("Set(<long-key>, bar) should return an error, got", err)
	}
}

func TestSQLKeyValueStore_UsingLongValueFails(t *testing.T) {
	kv := setUp(t)
	defer tearDown(db)
	defer kv.Close()

	// we limit the value length 8 characters in tests
	err := kv.Set("foo", "123456789")
	if err == nil {
		t.Error("Set(foo, <long-value>) should return an error, got", err)
	}
}

func setUpMySQL() (*sql.DB, error) {
	connectionString := os.Getenv("TEST_DB")
	if connectionString == "" {
		return nil, errors.New("No TEST_DB env variable set. Tip: use TEST_DB='user:pwd@tcp(localhost:3306)/test_db' go test")
	}
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("Cannot connect to the database: %v", err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS `test_config`")
	if err != nil {
		return nil, fmt.Errorf("Failed to drop table: %v", err)
	}

	_, err = db.Exec("CREATE TABLE `test_config` (" +
		"`key` VARCHAR(8) PRIMARY KEY," +
		"`value` VARCHAR(8) NOT NULL" +
		")")
	if err != nil {
		return nil, fmt.Errorf("Failed to create table: %v", err)
	}
	return db, nil
}

func tearDownMySQL() error {
	return db.Close()
}

func setUp(t *testing.T) *SQLKeyValueStore {
	kv, err := NewMySQLKeyValueStore(db, &SQLKeyValueStoreConfig{
		TableName:   "test_config",
		KeyColumn:   "key",
		ValueColumn: "value",
		MaxKeyLen:   8,
		MaxValueLen: 8,
	})
	if err != nil {
		t.Fatal("Cannot init SQLKeyValueStore:", err)
	}
	return kv
}


func tearDown(db *sql.DB) {
	_, err := db.Exec("TRUNCATE `test_config`")
	if err != nil {
		fmt.Print("Couldn't truncate test_config table")
	}
}
