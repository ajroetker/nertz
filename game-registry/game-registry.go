package main

import (
	"code.google.com/p/gcfg"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
)

type Config struct {
	Database struct {
		Dbname   string
		User     string
		Host     string
		Port     uint
		Password string
		Sslmode  string
	}
}

func main() {
	var cfg Config
	err := gcfg.ReadFileInto(&cfg, os.Args[1])

    dbcfg  := cfg.Database
	dbname := dbcfg.Dbname
	host   := dbcfg.Host
	port   := dbcfg.Port

	connstring := fmt.Sprintf("dbname=%v host=%v port=%v", dbname, host, port)
	db, err := sql.Open("postgres", connstring)
	println(db, err)

}
