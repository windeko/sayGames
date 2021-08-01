package logs

import (
	"encoding/json"
	"net/http"
	"strings"
)

func readLogsFromBody(r *http.Request) LogSlice {
	decoder := json.NewDecoder(r.Body)
	var t RawLogs
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	var res LogSlice
	for _, strLog := range strings.Split(t.Logs, "\n") {
		aLog := &Log{}
		json.Unmarshal([]byte(strLog), aLog)
		res = append(res, *aLog)
	}

	return res
}
