package logs

import (
	"sync"
	"time"
)

type Log struct {
	ClientTime string `json:"client_time"`
	DeviceId   string `json:"device_id"`
	DeviceOs   string `json:"device_os"`
	Session    string `json:"session"`
	Sequence   int    `json:"sequence"`
	Event      string `json:"event"`
	ParamInt   int    `json:"param_int"`
	ParamStr   string `json:"param_str"`
}

type RawLogs struct {
	Logs string `json:"logs"`
}

type EnrichedLogs struct {
	Log
	IP         string `json:"ip"`
	ServerTime string `json:"server_time"`
}

type LogSlice []Log

func (logSlice LogSlice) enrichLogs(remoteAddr string) []EnrichedLogs {
	wg := &sync.WaitGroup{}
	logNum := len(logSlice)
	bunch := make([]EnrichedLogs, 0, logNum)

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
	logChan := make(chan []EnrichedLogs, routineCount)

	// Пускаем в несколько горутин
	if logNum/defaultRoutineCount > 0 {
		for i := 0; i < defaultRoutineCount; i++ {
			wg.Add(1)
			go logSlice[i*logNum/defaultRoutineCount:(i+1)*logNum/defaultRoutineCount].enrichLog(logChan, remoteAddr, wg)
		}
	}

	// И еще одна горутина чтобы подбить остаток если он есть
	if logNum%defaultRoutineCount > 0 {
		wg.Add(1)
		go logSlice[logNum-(logNum%defaultRoutineCount):].enrichLog(logChan, remoteAddr, wg)
	}

	wg.Wait()

	for i := 0; i < routineCount; i++ {
		bunch = append(bunch, <-logChan...)
	}

	return bunch
}

func (logSlice LogSlice) enrichLog(logChan chan<- []EnrichedLogs, remoteAddr string, wg *sync.WaitGroup) {
	defer wg.Done()

	num := len(logSlice)

	bunch := make([]EnrichedLogs, 0, num)

	for _, aLog := range logSlice {
		curTime := time.Now()
		var eLog = EnrichedLogs{Log: aLog, ServerTime: curTime.Format("2006-01-02 15:04:05"), IP: remoteAddr}

		bunch = append(bunch, eLog)
	}

	logChan <- bunch
}
