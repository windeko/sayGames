package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mailru/go-clickhouse"
	"github.com/windeko/sayGames/log-generator/logs"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"sync"
	"time"
)

//type rawLogs struct {
//	Logs string `json:"logs"`
//}
//
//type logs struct {
//	ClientTime string `json:"client_time"`
//	DeviceId   string `json:"device_id"`
//	DeviceOs   string `json:"device_os"`
//	Session    string `json:"session"`
//	Sequence   int    `json:"sequence"`
//	Event      string `json:"event"`
//	ParamInt   int    `json:"param_int"`
//	ParamStr   string `json:"param_str"`
//}

//type enrichedLogs struct {
//	logs
//	IP         string `json:"ip"`
//	ServerTime string `json:"server_time"`
//}

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

	reqLogs := readLogsFromBody(r)

	mu.Lock()
	dump = append(dump, enrichLogs(reqLogs, r.RemoteAddr)...)
	fmt.Println(len(dump))
	if len(dump) > dumpThreshold {
		go saveToClickhouse(dump[:dumpThreshold])
		dump = dump[dumpThreshold:]
		fmt.Println(dumpThreshold, "logs are saved to DB")
	}
	mu.Unlock()

	duration := time.Since(start)
	fmt.Println(len(reqLogs), "logs are served for: ", duration)

	w.WriteHeader(http.StatusOK)
	return
}

func readLogsFromBody(r *http.Request) logs.LogSlice {
	decoder := json.NewDecoder(r.Body)
	var t logs.RawLogs
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	var res logs.LogSlice
	for _, strLog := range strings.Split(t.Logs, "\n") {
		aLog := &logs.Log{}
		json.Unmarshal([]byte(strLog), aLog)
		res = append(res, *aLog)
	}

	return res
}

func enrichLogs(logSlice logs.LogSlice, remoteAddr string) []logs.EnrichedLogs {
	wg := &sync.WaitGroup{}
	logNum := len(logSlice)
	bunch := make([]logs.EnrichedLogs, 0, logNum)

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
	logChan := make(chan []logs.EnrichedLogs, routineCount)

	// Пускаем в несколько горутин
	if logNum/defaultRoutineCount > 0 {
		for i := 0; i < defaultRoutineCount; i++ {
			wg.Add(1)
			go enrichLog(logChan, logSlice[i*logNum/defaultRoutineCount:(i+1)*logNum/defaultRoutineCount], remoteAddr, wg)
		}
	}

	// И еще одна горутина чтобы подбить остаток если он есть
	if logNum%defaultRoutineCount > 0 {
		wg.Add(1)
		go enrichLog(logChan, logSlice[logNum-(logNum%defaultRoutineCount):], remoteAddr, wg)
	}

	wg.Wait()

	for i := 0; i < routineCount; i++ {
		bunch = append(bunch, <-logChan...)
	}

	return bunch
}

func enrichLog(logChan chan<- []logs.EnrichedLogs, logSlice logs.LogSlice, remoteAddr string, wg *sync.WaitGroup) {
	defer wg.Done()

	num := len(logSlice)

	bunch := make([]logs.EnrichedLogs, 0, num)

	for _, aLog := range logSlice {
		curTime := time.Now()
		var eLog = logs.EnrichedLogs{Log: aLog, ServerTime: curTime.Format("2006-01-02 15:04:05"), IP: remoteAddr}

		bunch = append(bunch, eLog)
	}

	logChan <- bunch
}

func saveToClickhouse(enrichedLogs []logs.EnrichedLogs) {

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
