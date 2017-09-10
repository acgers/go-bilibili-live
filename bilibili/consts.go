package gbl

const (
	// get
	liveURL          = "http://live.bilibili.com/"
	signAPI          = "http://api.live.bilibili.com/sign/doSign"
	signInfoAPI      = "http://live.bilibili.com/sign/GetSignInfo"
	userInfoAPI      = "http://api.live.bilibili.com/User/getUserInfo"
	roomInfoAPI      = "http://live.bilibili.com/live/getInfo?roomid=%d"
	currentTaskAPI   = "http://live.bilibili.com/FreeSilver/getCurrentTask"
	dailyGiftAPI     = "http://api.live.bilibili.com/giftBag/sendDaily?_=%d"
	playerGiftBagAPI = "http://api.live.bilibili.com/gift/playerBag?_=%d"
	sendGiftAPI      = "http://api.live.bilibili.com/giftBag/send"

	// post
	heartAPI = "http://api.live.bilibili.com/User/userOnlineHeart"
)

const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.79 Safari/537.36"

const cookieDefault = `sid=9ciw7iqm; finger=14bc3c4e; DedeUserID=4535353;DedeUserID__ckMd5=2051c07989aa12c6; SESSDATA=f8ccd88e%2C1507272334%2Ce5f2be23; bili_jct=cc498b4cf4682ce68adcfe45dcad327e; biliMzIsnew=1; biliMzTs=150376330500; fts=1504434704681; LIVE_LOGIN_DATA=b4e07e2da07c58f3a930c604341000a84b5e7602990; LIVE_LOGIN_DATA__ckMd5=3adab4b634d4bd43319; LIVE_PLAYER_TYPE=2; buvid3=CB38288B-251F-4FDD-AF3D-FB6757B6C3341986infoc; _cnt_pm=0; _cnt_notify=4; LIVE_BUVID=ce8ec80c44cd618ef223587438908bba58; LIVE_BUVID__ckMd5=2710a8777d0f44838d;`

const envCookie = "GBL_COOKIE"
const envRoomID = "GBL_ROOMID"
