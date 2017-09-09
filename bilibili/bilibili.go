package gbl

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var (
	version     string
	built       string
	gitRevision string
)

var (
	debug  bool
	cookie string
	roomID int
)

// ParseFlag parse command args
func ParseFlag() {
	var printVersion bool

	flag.BoolVar(&printVersion, "v", false, "-v, print version")
	if printVersion {
		buildtime, _ := strconv.ParseInt(built, 0, 64)
		fmt.Println()
		fmt.Printf("go-bilibili-live version:           %s\n", version)
		fmt.Printf("go-bilibili-live git revision:      %s\n", gitRevision)
		fmt.Printf("go-bilibili-live build time:        %s\n", time.Unix(buildtime, 0))
		fmt.Printf("go runtime os:                      %s\n", runtime.GOOS)
		fmt.Printf("go runtime arch:                    %s\n", runtime.GOARCH)
		fmt.Printf("go runtime version:                 %s\n", runtime.Version())
		os.Exit(0)
	}

	cookie = os.Getenv(envCookie)
	roomID, _ = strconv.Atoi(os.Getenv(envRoomID))

	flag.BoolVar(&debug, "d", false, "-d=false or -d false, whether show debug log")
	if cookie == "" {
		flag.StringVar(&cookie, "c", cookieDefault,
			"-c=cookieValue or -c cookieValue, bilibili live cookie value")
	}
	if roomID == 0 {
		flag.IntVar(&roomID, "r", 320, "-r=320 or -r 320, up room id")
	}
	flag.Parse()

	if debug {
		debugln("cookie:", cookie, "roomId:", roomID)
	}
}

func loop() {
	defer recoverFunc()

	infoln("开始挂机...")

	wg := sync.WaitGroup{}
	wg.Add(1)

	r := getBilibili(userInfoAPI)
	var succeed bool
	code := r["code"]
	switch code.(type) {
	case string:
		succeed = code == "REPONSE_OK"
	case float64:
		succeed = code == 0
	}
	if !succeed {
		errorln("挂机失败:", r["msg"])
		os.Exit(1)
	}

	go sign()
	go onlineHeart()

	wg.Wait()
}

func sign() {
	doSign := func() bool {
		ret := getBilibili(fmt.Sprintf(dailyGiftAPI, time.Now().UnixNano()/1e6))
		if ret == nil || ret["code"].(float64) != 0 {
			errorln("获取每日礼物失败", ret["msg"])
			return false
		}
		ret = getBilibili(signAPI)
		if ret == nil || ret["code"].(float64) != 0 {
			errorln("签到失败:", ret["msg"])
			ret = getBilibili(signInfoAPI)
			if ret != nil {
				if ret = ret["data"].(map[string]interface{}); ret["status"].(float64) == 1 {
					infoln("今日已签到成功,", ret["text"], ",", ret["specialText"])
					return true
				}
			}
			return false
		}
		infoln("今日签到成功")
		return true
	}

	for {
		for {
			succeed := doSign()
			if succeed {
				break
			}
			time.Sleep(5 * time.Second)
		}

		now := time.Now()
		next := now.Add(24 * time.Hour)
		next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 40, 0, next.Location())
		infoln("下次签到时间:", next)
		timer := time.NewTimer(next.Sub(now))
		<-timer.C
	}
}

func onlineHeart() {
	heart := func() {
		ret := postBilibili(heartAPI, nil)
		if ret == nil || ret["code"].(float64) != 0 {
			errorln("心跳检测失败:", ret["msg"])
		} else {
			infoln("心跳检测成功")
		}
	}
	heart()
	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-ticker.C:
			updateSettingsFromEnv()
			go heart()
		}
	}
}
