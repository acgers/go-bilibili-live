package gbl

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/gomail.v2"
)

type bilibiliResp = map[string]interface{}

type header = struct {
	name  string
	value string
}

type urlValues = url.Values

var recoverFunc = func() {
	if rec := recover(); rec != nil {
		errorln(rec)
	}
}

var mailClient *gomail.Dialer

func getBilibili(url string, headers ...header) bilibiliResp {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errorln(err)
		return nil
	}

	return requestBilibili(req, headers...)
}

func postBilibili(url string, postBody interface{}, headers ...header) bilibiliResp {
	var req *http.Request
	var err error

	switch postBody.(type) {
	case urlValues:
		req, err = http.NewRequest("POST", url, strings.NewReader(postBody.(urlValues).Encode()))
	default:
		if postBody != nil {
			var jsonBody []byte
			jsonBody, err = json.Marshal(postBody)
			if err != nil {
				errorln(err)
				return nil
			}
			req, err = http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
		} else {
			req, err = http.NewRequest("POST", url, nil)
		}
	}

	if err != nil {
		errorln(err)
		return nil
	}

	return requestBilibili(req, headers...)
}

func requestBilibili(req *http.Request, headers ...header) (data bilibiliResp) {
	req.Header.Set("Referer", fmt.Sprintf("http://live.bilibili.com/%d", roomID))
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	for _, header := range headers {
		req.Header.Set(header.name, header.value)
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		errorln(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorln(err)
		return
	}

	if resp.Header.Get("Content-Encoding") == "gzip" {
		var reader *gzip.Reader
		buf := bytes.NewBuffer(body)
		reader, err = gzip.NewReader(buf)
		if err != nil {
			errorln(err)
			return
		}
		defer reader.Close()

		dec := json.NewDecoder(reader)
		err = dec.Decode(&data)
		if err != nil {
			errorln(err)
			return
		}
	} else {
		err = json.Unmarshal(body, &data)
		if err != nil {
			errorln(err)
			return
		}
	}

	if debug {
		debugln("%v\n", data)
	}

	return
}

func updateSettingsFromEnv() {
	ck := os.Getenv(envCookie)
	rmi := os.Getenv(envRoomID)
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

func checkMailSetting() {
	if mailClient == nil {
		if gblMail == "" {
			panicln("GBL_MAIL flags not set")
		}
		if gblPwd == "" {
			panicln("GBL_MAIL_PWD flags not set")
		}
		if gblSMTP == "" {
			panicln("GBL_MAIL_SMTP flags not set")
		}
		mailClient = gomail.NewDialer(gblSMTP, 465, gblMail, gblPwd)
	}
}

func sendMail() {
	message := gomail.NewMessage()
	message.SetHeader("From", gblMail)
	message.SetHeader("To", notifyMail)
	message.SetHeader("Subject", "[GBL] bilibili live 任务运行失败")
	message.SetBody("text/html; charset=utf-8",
		fmt.Sprintf("<pre>请更新Cookie,可以更新环境变量[%s]的值</pre>", envCookie))
	err := mailClient.DialAndSend(message)
	if err != nil {
		errorln(err)
	}
}

func isEmail(str string) bool {
	ret, _ := regexp.MatchString(`^[a-z0-9A-Z]+([\-_\.][a-z0-9A-Z]+)*@([a-z0-9A-Z]+(-[a-z0-9A-Z]+)*\.)+[a-zA-Z]+$`, str)
	return ret
}
