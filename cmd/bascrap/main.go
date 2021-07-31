package main

import (
	"github.com/posipaka-trade/bascrap/worker"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
)

func main() {
	cmn.InitLoggers("bascrap")
	worker.StartMonitoring()
}
