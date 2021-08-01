package logs

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
