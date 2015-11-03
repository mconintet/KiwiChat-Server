## About
It's server side for [KiwiChat](https://github.com/mconintet/KiwiChat). It use MySQL as it's data storage, only stores some basic info of users and their friend relationship. **Doesn't store any message sending between users.**

## Installation
1. Install dependencies at first:

  ```go
  go get github.com/go-sql-driver/mysql
  go get github.com/bitly/go-simplejson
  go get github.com/mconintet/kiwi
  ```

2. Download source code and import the `db.sql` under source code directory into you MySQL database.

3. You can build source code optionally or just use `go run`.

## Usage

```go
Usage:
  -db string
    	database access info
  -sa string
    	server address (default ":9876")
```

```go
// example
go run *.go -db="root:pass@unix(/tmp/mysql.sock)/kiwi"
``` 

More details about the format of db access string please see [dsn-data-source-name](https://github.com/go-sql-driver/mysql#dsn-data-source-name).