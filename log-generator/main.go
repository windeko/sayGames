package main

import (
	"github.com/windeko/sayGames/log-generator/logs"
)

func main() {
	for {
		logs.GenerateAndSendLogs(30)
	}
}
