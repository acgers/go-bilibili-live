package gbl

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	debugln func(v ...interface{})
	infoln  func(v ...interface{})
	errorln func(v ...interface{})
	panicln func(v ...interface{})
)

func initLogger() {
	logPath := filepath.Join(os.TempDir(), string(os.PathSeparator),
		fmt.Sprintf("dpl.%s.log", time.Now().Format("20060102150405")))
	log.Println("log file path:", logPath)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	logger := log.New(io.MultiWriter(os.Stdout, logFile), "[gbl]: ", log.LstdFlags)

	debugln = func(v ...interface{}) {
		logger.Println("[debug]", v)
	}

	infoln = func(v ...interface{}) {
		logger.Println("[info]", v)
	}

	errorln = func(v ...interface{}) {
		logger.Println("[error]", v)
	}

	panicln = func(v ...interface{}) {
		logger.Panicln("[panic]", v)
	}
}
