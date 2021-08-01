package logs

import (
	"math/rand"
	"sync"
	"time"
)

var gameEvents = [...]string{"app_start", "having_fun", "getting_bored", "donate", "mining", "app_end"}

func makeBunch(logChan chan<- []Log, num int, wg *sync.WaitGroup) {
	defer wg.Done()

	bunch := make([]Log, 0, num)

	for i := 0; i < num; i++ {
		bunch = append(bunch, makeLog())
	}

	logChan <- bunch
}

func makeLog() Log {
	curTime := time.Now()
	return Log{
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
