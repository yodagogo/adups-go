package main

import (
	"adups-go/logger"

	"go.uber.org/zap"
)

//Log xx
var Log *zap.Logger

//Sugger xx
var Sugger *zap.SugaredLogger

func main() {
	cfg := logger.Config{Mode: "detail", LogPath: "/tmp/test.log"}
	Log, Sugger = cfg.BuildConfig()
	Log.Info("Logger,hello world")
	Sugger.Info("Sugger,hello world")
}
