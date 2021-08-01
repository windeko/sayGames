package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mailru/go-clickhouse"
	"github.com/windeko/sayGames/log-receiver/logs"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"
)

var DBConnect *sql.DB
var dump = make([]logs.EnrichedLogs, 0, 6000)
var dumpThreshold int = 5000
var mu = sync.Mutex{}

func main() {
	var err error
	DBConnect, err = sql.Open("clickhouse", "http://clickhouseServer:8123/sayGames")
	//DBConnect, err = sql.Open("clickhouse", "http://localhost:8123/sayGames")
	if err != nil {
		log.Fatal(err)
	}
	if err := DBConnect.Ping(); err != nil {
		log.Fatal(err)
	}

	c := &logs.Context{DB: DBConnect}

	http.HandleFunc("/logs", c.ServeLogs)

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}
