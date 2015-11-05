[![Coverage Status](https://coveralls.io/repos/petrkotek/gkv/badge.svg?branch=master&service=github)](https://coveralls.io/github/petrkotek/gkv?branch=master)

# gkv
Extremely simple Key/Value store for Go (golang).

Currently comes with only one backend implementation - MySQL. `gkv` doesn't support transactions or other fancy features.

The only methods it supports are `Set(key, value)`, `Get(key)` and `Delete(key)`, where both `key` and `value` are strings. 
 
## Example Usage
#### 1. Create database table on your MySQL server, e.g.:
```sql
CREATE TABLE `key_value_store`
  `key` VARBINARY(32) PRIMARY KEY,
  `value` VARBINARY(128) NOT NULL
)
```

#### 2. Import `gkv` in your `.go` file:
```go
import "github.com/petrkotek/gkv"
```

#### 3. Use `gkv`!
```go
db, err := sql.Open("mysql", "user:pwd@tcp(localhost:3306)/test_db")
if err != nil {
    // TODO: handle error
}

kv, err := gkv.NewMySQLKeyValueStore(db, &gkv.SQLKeyValueStoreConfig{
    TableName:   "key_value_store",
    KeyColumn:   "key",
    ValueColumn: "value",
    MaxKeyLen:   32,
    MaxValueLen: 128,
})
if err != nil {
    // TODO: handle error
}

// set a value for the key
err = kv.Set("foo", "bar")
if err != nil {
    // TODO: handle error
}

// retrieve the value
value, err := kv.Get("foo")
if err != nil {
    // TODO: handle error
}
fmt.Print(value)

// delete the key
err = kv.Del("foo")
if err != nil {
    // TODO: handle error
}

kv.Close()
```

## Other Key/Value Stores for Go
This project is useful only for limited number of use cases (such as when you already have MySQL database/cluster running and don't want to use proper Key-Value database).
 
In many cases, safer choice will be using for instance one of these packages, which purely use Go and file-backed storage:

* [boltdb/bolt](https://github.com/boltdb/bolt)
* [steveyen/gkvlite](https://github.com/steveyen/gkvlite)
