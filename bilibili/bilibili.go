package gbl

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"os"
	"regexp"
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
	gblMail string
	gblPwd  string
	gblSMTP string
)

var (
	debug      bool
	cookie     string
	roomID     int
	notifyMail string
	roomUID    int
	upNickName string
	token      string
)

// ParseFlag parse command args
func ParseFlag() {
	cookie = os.Getenv(envCookie)
	roomID, _ = strconv.Atoi(os.Getenv(envRoomID))

	var printVersion bool
	var cmdCookie, cmdNotifyMail string
	var cmdRoomID int
	flag.BoolVar(&printVersion, "v", false, "-v, print version")
	flag.BoolVar(&debug, "d", false, "-d=false, whether show debug log")
	flag.StringVar(&cmdCookie, "c", cookieDefault, "-c=cookieValue, bilibili live cookie value")
	flag.IntVar(&cmdRoomID, "r", 320, "-r=320, up room id")
	flag.StringVar(&cmdNotifyMail, "m", "", "-m=a@b.c, mail for notify")

	flag.Parse()

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

	initLogger()

	if cookie == "" {
		cookie = cmdCookie
	}
	if roomID == 0 {
		roomID = cmdRoomID
	}

	if cmdNotifyMail != "" {
		if !isEmail(cmdNotifyMail) {
			panicln("非法参数，不是合法的邮件地址")
		}
		notifyMail = cmdNotifyMail
		checkMailSetting()
	}

	if debug {
		debugln("cookie:", cookie, "roomId:", roomID)
	}
}

func loop() {
	infoln("开始挂机...")

	wg := sync.WaitGroup{}
	wg.Add(1)

	rand.Seed(time.Now().UnixNano())

	resp, err := http.Get(liveURL + strconv.Itoa(roomID))
	if err != nil {
		panicln(err)
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	htmlBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panicln(err)
	}
	roomIDRegexp := regexp.MustCompile(`var ROOMID = (\d+)`)
	matchs := roomIDRegexp.FindStringSubmatch(string(htmlBody))
	if debug {
		debugln(matchs)
	}

	if len(matchs) != 2 || matchs[1] == "" {
		panicln("获取房间[", roomID, "]cid失败")
	}

	roomID, err = strconv.Atoi(matchs[1])
	if err != nil {
		panicln(err)
	}

	tokenRegexp := regexp.MustCompile(`LIVE_LOGIN_DATA=(.{40})`)
	matchs = tokenRegexp.FindStringSubmatch(cookie)
	if debug {
		debugln(matchs)
	}
	if len(matchs) != 2 || matchs[1] == "" {
		errorln("当前Cookie可能不是在live.bilibili.com获取的")
	} else {
		token = matchs[1]
	}

	ret := getBilibili(userInfoAPI)
	var succeed bool
	code := ret["code"]
	switch code.(type) {
	case string:
		succeed = code == "REPONSE_OK"
	case float64:
		succeed = code == 0
	}
	if !succeed {
		errorln("挂机失败:", ret["msg"])
		os.Exit(1)
	}

	sign()
	userOnlineHeart()
	sendOutdatedGift()
	openSilverBox()

	wg.Wait()
}

func openSilverBox() {

}

func sign() {
	doSign := func() bool {
		defer recoverFunc()
		ret := getBilibili(fmt.Sprintf(dailyGiftAPI, time.Now().UnixNano()/1e6))
		if ret == nil || ret["code"].(float64) != 0 {
			errorln("获取每日礼物失败", ret["msg"])
			return false
		}

		ret = getBilibili(signAPI)
		if ret == nil || ret["code"].(float64) != 0 {
			errorln("签到失败:", ret["msg"])
		}
		ret = getBilibili(signInfoAPI)
		if ret != nil {
			data := ret["data"].(map[string]interface{})
			if data["status"].(float64) == 1 {
				infoln("今天已签到成功,", data["text"], ",", data["specialText"])
				return true
			}
		}
		return false
	}

	go func() {
		for {
			for {
				if succeed := doSign(); succeed {
					break
				}
				errorln("每日签到失败，重试...")
				time.Sleep(10 * time.Second)
			}

			now := time.Now()
			next := now.Add(24 * time.Hour)
			next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 40, 0, next.Location())
			infoln("下次签到时间:", next)
			timer := time.NewTimer(next.Sub(now))
			<-timer.C
		}
	}()
}

func userOnlineHeart() {
	onlineHeart := func() {
		defer recoverFunc()
		ret := postBilibili(heartAPI, nil)
		if ret == nil || ret["code"].(float64) != 0 {
			errorln("心跳检测失败:", ret["msg"])
		} else {
			infoln("心跳检测成功")
		}
	}

	go func() {
		onlineHeart()
		ticker := time.NewTicker(5 * time.Minute)
		for {
			select {
			case <-ticker.C:
				updateSettingsFromEnv()
				onlineHeart()
			}
		}
	}()
}

func sendOutdatedGift() {
	sendGift := func() bool {
		defer recoverFunc()
		ret := getBilibili(fmt.Sprintf(playerGiftBagAPI, time.Now().UnixNano()/1e6))
		if ret == nil || ret["code"].(float64) != 0 {
			errorln("获取礼物包裹失败", ret["msg"])
			return false
		}
		debugln(ret)
		datas := ret["data"].([]interface{})
		for _, data := range datas {
			gift := data.(map[string]interface{})
			if gift["expireat"] == "今日" {
				if roomUID == 0 {
					ret = getBilibili(fmt.Sprintf(roomInfoAPI, roomID))
					if ret == nil || ret["code"].(float64) != 0 {
						errorln("获取房间信息失败", ret["msg"])
						return false
					}
					data := ret["data"].(map[string]interface{})
					roomUID = int(data["MASTERID"].(float64))
					upNickName = data["ANCHOR_NICK_NAME"].(string)
				}
				body := make(urlValues, 9)
				body.Set("giftId", strconv.FormatInt(int64(gift["gift_id"].(float64)), 10))
				body.Set("roomid", strconv.Itoa(roomID))
				body.Set("ruid", strconv.Itoa(roomUID))
				body.Set("num", strconv.Itoa(int(gift["gift_num"].(float64))))
				// test
				// body.Set("num", strconv.Itoa(1))
				body.Set("coinType", "silver")
				body.Set("Bag_id", strconv.FormatInt(int64(gift["id"].(float64)), 10))
				body.Set("timestamp", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
				body.Set("rnd", strconv.Itoa(int(rand.Int31n(math.MaxInt32))))
				body.Set("token", token)
				ret = postBilibili(sendGiftAPI, body,
					header{"Content-Type", "application/x-www-form-urlencoded; charset=UTF-8"})
				if ret == nil || ret["code"].(float64) != 0 {
					errorln("赠送礼物失败", ret["msg"])
					return false
				}
				infoln("礼物投喂成功 [to:", upNickName, "count:", gift["gift_num"],
					"name", gift["gift_name"], "]")
			}
		}
		return true
	}

	go func() {
		for {
			for {
				if succeed := sendGift(); succeed {
					break
				}
				errorln("赠送过期礼物失败，重试...")
				time.Sleep(10 * time.Second)
			}

			now := time.Now()
			next := now.Add(24 * time.Hour)
			next = time.Date(next.Year(), next.Month(), next.Day(), 23, 50, 0, 0, next.Location())
			infoln("下次赠送时间:", next)
			timer := time.NewTimer(next.Sub(now))
			<-timer.C
		}
	}()
}
