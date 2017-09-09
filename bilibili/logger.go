package gbl

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var logger *log.Logger

func initLogger() {
	logPath := filepath.Join(os.TempDir(), string(os.PathSeparator),
		fmt.Sprintf("dpl.%s.log", time.Now().Format("20060102150405")))
	log.Println("log file path:", logPath)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	logger = log.New(io.MultiWriter(os.Stdout, logFile), "[gbl]: ", log.LstdFlags)
}
