package main

import (
    "code.google.com/p/gcfg"
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"
    "os"
    "log"
)

func checkErr(err error, msg string) {
    if err != nil {
        log.Fatalln(msg, err)
    }
}

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

func (cfg Config) ConnString() string {
    dbcfg := cfg.Database
    connstr := fmt.Sprintf("dbname=%v%%v", dbcfg.Dbname)
    if dbcfg.User != "" {
        tmp := fmt.Sprintf(" user=%v%%v", dbcfg.User)
        connstr = fmt.Sprintf(connstr, tmp)
    }
    if dbcfg.Host != "" {
        tmp := fmt.Sprintf(" host=%v%%v", dbcfg.Host)
        connstr = fmt.Sprintf(connstr, tmp)
    }
    if dbcfg.Port != 0 {
        tmp := fmt.Sprintf(" port=%v%%v", dbcfg.Port)
        connstr = fmt.Sprintf(connstr, tmp)
    }
    if dbcfg.Password != "" {
        tmp := fmt.Sprintf(" password='%v'%%v", dbcfg.Password)
        connstr = fmt.Sprintf(connstr, tmp)
    }
    if dbcfg.Sslmode == "disable" || dbcfg.Sslmode == "require" || dbcfg.Sslmode == "verify-full" {
        tmp := fmt.Sprintf(" sslmode=%v%%v", dbcfg.Sslmode)
        connstr = fmt.Sprintf(connstr, tmp)
    }
    connstr = fmt.Sprintf(connstr, "")
    return connstr
}

func main() {
    var cfg Config
    err := gcfg.ReadFileInto(&cfg, os.Args[1])
    checkErr(err, "Reading gcfg file failed")

    connstring := cfg.ConnString()
    db, err := sql.Open("postgres", connstring)
    checkErr(err, "Opening database connection failed")
    defer db.Close()

    fmt.Printf("Connected to postgresql with:\n    %v\n", connstring)
    fmt.Printf("%v\n", db)

}
