package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const verson = "0.9.8"

const (
	// get
	sign        = "http://api.live.bilibili.com/sign/doSign"
	signInfo    = "http://live.bilibili.com/sign/GetSignInfo"
	userInfo    = "http://api.live.bilibili.com/User/getUserInfo"
	currentTask = "http://live.bilibili.com/FreeSilver/getCurrentTask"

	// post
	heart = "http://api.live.bilibili.com/User/userOnlineHeart"
)

const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.79 Safari/537.36"

const cookieDefault = `sid=9ciw7iqm; finger=14bc3c4e; DedeUserID=4535353;DedeUserID__ckMd5=2051c07989aa12c6; SESSDATA=f8ccd88e%2C1507272334%2Ce5f2be23; bili_jct=cc498b4cf4682ce68adcfe45dcad327e; biliMzIsnew=1; biliMzTs=150376330500; fts=1504434704681; LIVE_LOGIN_DATA=b4e07e2da07c58f3a930c604341000a84b5e7602990; LIVE_LOGIN_DATA__ckMd5=3adab4b634d4bd43319; LIVE_PLAYER_TYPE=2; buvid3=CB38288B-251F-4FDD-AF3D-FB6757B6C3341986infoc; _cnt_pm=0; _cnt_notify=4; LIVE_BUVID=ce8ec80c44cd618ef223587438908bba58; LIVE_BUVID__ckMd5=2710a8777d0f44838d;`

var logger *log.Logger

type bilibiliResp = map[string]interface{}

func init() {
	logPath := filepath.Join(os.TempDir(), string(os.PathSeparator),
		fmt.Sprintf("dpl.%s.log", time.Now().Format("20060102150405")))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	logger = log.New(io.MultiWriter(os.Stdout, logFile), "[gbl]:", log.LstdFlags)

	handlerSignal()
}

var debug bool
var printVersion bool
var cookie string
var roomID int

func main() {
	parseFlag()

	logger.Println("开始挂机...")

	wg := sync.WaitGroup{}
	wg.Add(1)

	r := getBilibili(userInfo)
	var succeed bool
	code := r["code"]
	switch code.(type) {
	case string:
		succeed = code == "REPONSE_OK"
	case float64:
		succeed = code == 0
	}
	if !succeed {
		logger.Println("挂机失败:", r["msg"])
		os.Exit(1)
	}
	getBilibili(sign)
	getBilibili(signInfo)
	go onlineHeart()

	wg.Wait()
}

func onlineHeart() {
	heart := func() {
		r := postBilibili(heart, nil)
		if r == nil || r["code"].(float64) != 0 {
			logger.Println("心跳检测失败:", r["msg"])
		} else {
			logger.Println("心跳检测成功")
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

func getBilibili(url string) bilibiliResp {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Println(err)
		return nil
	}

	return requestBilibili(req)
}

func postBilibili(url string, postBody interface{}) bilibiliResp {
	var req *http.Request
	var err error

	if postBody != nil {
		var jsonBody []byte
		jsonBody, err = json.Marshal(postBody)
		if err != nil {
			logger.Println(err)
			return nil
		}
		req, err = http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	} else {
		req, err = http.NewRequest("POST", url, nil)
	}

	if err != nil {
		logger.Println(err)
		return nil
	}

	return requestBilibili(req)
}

func requestBilibili(req *http.Request) (data bilibiliResp) {
	req.Header.Set("Referer", fmt.Sprintf("http://live.bilibili.com/%d", roomID))
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		logger.Println(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Println(err)
		return
	}

	if resp.Header.Get("Content-Encoding") == "gzip" {
		var reader *gzip.Reader
		buf := bytes.NewBuffer(body)
		reader, err = gzip.NewReader(buf)
		if err != nil {
			logger.Println(err)
			return
		}
		defer reader.Close()

		dec := json.NewDecoder(reader)
		err = dec.Decode(&data)
		if err != nil {
			logger.Println(err)
			return
		}
	} else {
		err = json.Unmarshal(body, &data)
		if err != nil {
			logger.Println(err)
			return
		}
	}

	if debug {
		logger.Printf("%v\n", data)
	}

	return
}

func parseFlag() {
	cookie = os.Getenv("GBL_COOKIE")
	roomID, _ = strconv.Atoi(os.Getenv("GBL_ROOMID"))

	flag.BoolVar(&debug, "d", false, "-d=false, whether show debug log")
	flag.BoolVar(&printVersion, "v", false, "-v, print version")
	if cookie == "" {
		flag.StringVar(&cookie, "c", cookieDefault, "-c=cookieValue, bilibili live cookie value")
	}
	if roomID == 0 {
		flag.IntVar(&roomID, "r", 320, "-r=320, up room id")
	}
	flag.Parse()

	if printVersion {
		fmt.Printf("go-bilibili-live Version: %s\n", verson)
		fmt.Printf("go runtime Version:       %s\n", runtime.Version())
		fmt.Printf("go runtime Arch:          %s\n", runtime.GOARCH)
		fmt.Printf("go runtime OS:            %s\n", runtime.GOOS)
		os.Exit(0)
	}

	if debug {
		logger.Println("cookie:", cookie, "roomId:", roomID)
	}
}

func updateSettingsFromEnv() {
	ck := os.Getenv("GBL_COOKIE")
	rmi := os.Getenv("GBL_ROOMID")
	if ck != "" {
		cookie = ck
	}
	if rmi != "" {
		id, err := strconv.Atoi(rmi)
		if err == nil {
			roomID = id
		}
	}
}
