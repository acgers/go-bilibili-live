package gbl

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type bilibiliResp = map[string]interface{}

var recoverFunc = func() {
	if rec := recover(); rec != nil {
		errorln(rec)
	}
}

func debugln(v ...interface{}) {
	logger.Println("[debug]", v)
}

func infoln(v ...interface{}) {
	logger.Println("[info]", v)
}

func errorln(v ...interface{}) {
	logger.Println("[error]", v)
}

func getBilibili(url string) bilibiliResp {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errorln(err)
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
			errorln(err)
			return nil
		}
		req, err = http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	} else {
		req, err = http.NewRequest("POST", url, nil)
	}

	if err != nil {
		errorln(err)
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
