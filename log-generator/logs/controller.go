package logs

import (
	"fmt"
	"sync"
	"time"
)

func GenerateAndSendLogs(logNum int) {
	mocLogs := createLogs(logNum)
	rawMocLogs := mocLogs.prepareLogReqBody()
	rawMocLogs.sendRequest()
}

func createLogs(logNum int) LogSlice {
	start := time.Now()

	wg := &sync.WaitGroup{}
	bunch := make([]Log, 0, logNum)

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
	logChan := make(chan []Log, routineCount)

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
