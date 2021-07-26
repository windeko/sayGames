package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type log struct {
	ClientTime string `json:"client_time"`
	DeviceId   string `json:"device_id"`
	DeviceOs   string `json:"device_os"`
	Session    string `json:"session"`
	Sequence   int    `json:"sequence"`
	Event      string `json:"event"`
	ParamInt   int    `json:"param_int"`
	ParamStr   string `json:"param_str"`
}

type rawLogs struct {
	Logs string `json:"logs"`
}

var gameEvents = [...]string{"app_start", "having_fun", "getting_bored", "donate", "mining", "app_end"}

func main() {
	for {
		time.Sleep(2 * time.Second)
		sendLogs(createLogs(10000))
	}
}

func createLogs(logNum int) []log {
	start := time.Now()

	wg := &sync.WaitGroup{}
	bunch := make([]log, 0, logNum)

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
	logChan := make(chan []log, routineCount)

	// Пускаем в несколько горутин
	if logNum/defaultRoutineCount > 0 {
		for i := 0; i < defaultRoutineCount; i++ {
			wg.Add(1)
			go makeBunch(logChan, logNum/defaultRoutineCount, wg)
		}
	}
	// И еще одна горутина чтобы подбить остаток если он есть
	if logNum%defaultRoutineCount > 0 {
		wg.Add(1)
		go makeBunch(logChan, logNum%defaultRoutineCount, wg)
	}

	wg.Wait()

	for i := 0; i < routineCount; i++ {
		bunch = append(bunch, <-logChan...)
	}

	duration := time.Since(start)
	fmt.Println(logNum, "MADE FOR", duration)

	return bunch
}

func makeBunch(logChan chan<- []log, num int, wg *sync.WaitGroup) {
	defer wg.Done()

	bunch := make([]log, 0, num)

	for i := 0; i < num; i++ {
		bunch = append(bunch, makeLog())
	}

	logChan <- bunch
}

func makeLog() log {
	curTime := time.Now()
	return log{
		ClientTime: curTime.Format("2006-01-02 15:04:05"),
		DeviceId:   "iPhone",
		DeviceOs:   "iOS 10",
		Session:    "anySession",
		Sequence:   rand.Int(),
		Event:      gameEvents[rand.Intn(len(gameEvents))],
		ParamInt:   rand.Int(),
		ParamStr:   "Auto Generated Log",
	}
}

func sendLogs(logs []log) {
	logsStrBuffer := make([]string, 0, len(logs))
	for _, log := range logs {
		jsonLog, _ := json.Marshal(&log)
		logsStrBuffer = append(logsStrBuffer, string(jsonLog))
	}

	logsString := strings.Join(logsStrBuffer, "\n")
	sendLogsRequest(logsString)
}

func sendLogsRequest(logs string) bool {
	url := "http://logReceiver:8080/logs"

	rawLogs := rawLogs{Logs: logs}
	rawLogsBytes, _ := json.Marshal(rawLogs)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(rawLogsBytes))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return true
}
