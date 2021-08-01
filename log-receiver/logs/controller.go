package logs

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Context struct {
	DB *sql.DB
}

var dump = make([]EnrichedLogs, 0, 6000)
var dumpThreshold int = 5000
var mu = sync.Mutex{}

func (c Context) ServeLogs(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	reqLogs := readLogsFromBody(r)

	mu.Lock()
	dump = append(dump, reqLogs.enrichLogs(r.RemoteAddr)...)
	fmt.Println(len(dump))
	if len(dump) > dumpThreshold {
		go c.saveToClickhouse(dump[:dumpThreshold])
		dump = dump[dumpThreshold:]
		fmt.Println(dumpThreshold, "logs are saved to DB")
	}
	mu.Unlock()

	duration := time.Since(start)
	fmt.Println(len(reqLogs), "logs are served for: ", duration)

	w.WriteHeader(http.StatusOK)
	return
}

func (c Context) saveToClickhouse(enrichedLogs []EnrichedLogs) {
	// Making transaction
	tx, err := c.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Creating request blueprint
	stmt, err := tx.Prepare(`
		INSERT INTO logs (
			clientTime,
			deviceId,
			deviceOs,
			session,
			sequence,
			event,
		    paramInt,
			paramStr,
			ip,
			serverTime
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)`)
	if err != nil {
		log.Fatal(err)
	}

	// Making requests
	for _, enLog := range enrichedLogs {
		if _, err := stmt.Exec(
			enLog.ClientTime,
			enLog.DeviceId,
			enLog.DeviceOs,
			enLog.Session,
			enLog.Sequence,
			enLog.Event,
			enLog.ParamInt,
			enLog.ParamStr,
			enLog.IP,
			enLog.ServerTime,
		); err != nil {
			log.Fatal(err)
		}
	}

	// Commit them
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

}
