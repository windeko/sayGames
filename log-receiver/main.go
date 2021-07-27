package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mailru/go-clickhouse"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"sync"
	"time"
)

type rawLogs struct {
	Logs string `json:"logs"`
}

type logs struct {
	ClientTime string `json:"client_time"`
	DeviceId   string `json:"device_id"`
	DeviceOs   string `json:"device_os"`
	Session    string `json:"session"`
	Sequence   int    `json:"sequence"`
	Event      string `json:"event"`
	ParamInt   int    `json:"param_int"`
	ParamStr   string `json:"param_str"`
}

type enrichedLogs struct {
	logs
	IP         string `json:"ip"`
	ServerTime string `json:"server_time"`
}

var DBConnect *sql.DB

func main() {
	var err error
	//DBConnect, err = sql.Open("clickhouse", "http://clickhouse_server:8123/sayGames")
	DBConnect, err = sql.Open("clickhouse", "http://localhost:8123/sayGames")
	if err != nil {
		log.Fatal(err)
	}
	if err := DBConnect.Ping(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/logs", serveLogs)

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}

func serveLogs(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	logs := readLogsFromBody(r)
	enrichedLogs := enrichLogs(logs, r.RemoteAddr)

	go saveToClickhouse(enrichedLogs)

	duration := time.Since(start)
	fmt.Println(len(enrichedLogs), "logs are served for: ", duration)

	w.WriteHeader(http.StatusOK)
	return
}

func readLogsFromBody(r *http.Request) []logs {
	decoder := json.NewDecoder(r.Body)
	var t rawLogs
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	var res []logs
	for _, strLog := range strings.Split(t.Logs, "\n") {
		log := &logs{}
		json.Unmarshal([]byte(strLog), log)
		res = append(res, *log)
	}

	return res
}

func enrichLogs(logs []logs, remoteAddr string) []enrichedLogs {
	wg := &sync.WaitGroup{}
	logNum := len(logs)
	bunch := make([]enrichedLogs, 0, logNum)

	// Количество используемых горутин. Впринципе можно вынести в environment
	defaultRoutineCount, routineCount := 5, 5
	// Если количество формируемых логов меньше чем количество Горутин - то мы схватим дедлок
	if logNum/defaultRoutineCount == 0 {
		routineCount = 0
	}
	if logNum%defaultRoutineCount > 0 {
		routineCount++
	}

	// Делаем буферизированный канал в который будем писать из нескольких рутин
	logChan := make(chan []enrichedLogs, routineCount)

	// Пускаем в несколько горутин
	if logNum/defaultRoutineCount > 0 {
		for i := 0; i < defaultRoutineCount; i++ {
			wg.Add(1)
			go enrichLog(logChan, logs[i*logNum/defaultRoutineCount:(i+1)*logNum/defaultRoutineCount], remoteAddr, wg)
		}
	}

	// И еще одна горутина чтобы подбить остаток если он есть
	if logNum%defaultRoutineCount > 0 {
		wg.Add(1)
		go enrichLog(logChan, logs[logNum-(logNum%defaultRoutineCount):], remoteAddr, wg)
	}

	wg.Wait()

	for i := 0; i < routineCount; i++ {
		bunch = append(bunch, <-logChan...)
	}

	return bunch
}

func enrichLog(logChan chan<- []enrichedLogs, logs []logs, remoteAddr string, wg *sync.WaitGroup) {
	defer wg.Done()

	num := len(logs)

	bunch := make([]enrichedLogs, 0, num)

	for _, log := range logs {
		curTime := time.Now()
		var eLog = enrichedLogs{logs: log, ServerTime: curTime.Format("2006-01-02 15:04:05"), IP: remoteAddr}

		bunch = append(bunch, eLog)
	}

	logChan <- bunch
}

func saveToClickhouse(enrichedLogs []enrichedLogs) {

	// Making transaction
	tx, err := DBConnect.Begin()
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
