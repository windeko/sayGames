package logs

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
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

type LogSlice []Log

func (logs LogSlice) prepareLogReqBody() RawLogs {
	logsStrBuffer := make([]string, 0, len(logs))
	for _, log := range logs {
		jsonLog, _ := json.Marshal(&log)
		logsStrBuffer = append(logsStrBuffer, string(jsonLog))
	}

	logsStr := strings.Join(logsStrBuffer, "\n")
	return RawLogs{Logs: logsStr}
}

func (rawLogs RawLogs) sendRequest() bool {
	url := "http://logReceiver:8080/logs"
	//url := "http://localhost:8080/logs"

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
