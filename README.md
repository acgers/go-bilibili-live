# BiliBili live 挂机的一些辅助功能

## Build Status
[![Build Status](https://travis-ci.org/acgers/go-bilibili-live.svg?branch=master)](https://travis-ci.org/acgers/go-bilibili-live)

## Usage

* ### 命令行传参

```
    gbl_[darwin_amd64|linux_386|linux_amd64|windows_386.exe|windows_amd64.exe]

    -h                         获取帮助信息

    -d <debug>                 true/false 是否打印debug日志

    -c <cookieValue>           bilibili live 的浏览器Cookie值

    -r <roomId>                up主直播间的房间号(用于自动投喂即将过期的礼物)

    -m <notifyMail>            接收通知的邮件地址

    -v                         打印版本信息
```

* ### 环境变量传参(优先取环境变量的参数)
bash/sh
```
export GBL_COOKIE="cookie_value"
export GBL_ROOMID=320
```

windows

在环境变量中添加新值

## bilibili live Cookie
需要切换到html播放器获取，有效期似乎长一些

## 例子(注意双引号)
```
gbl -d=false -r=320 -m="a@b.c" -c="sid=9ciw7iqm; DedeUserID=4535353 省略1000字;"
```

## 下载
[Download](https://github.com/acgers/go-bilibili-live/releases)

## 从源码安装(需要go1.9+环境)
```
go get github.com/acgers/go-bilibili-live
cd $GOPATH/src/github.com/acgers/go-bilibili-live
make && make install
```
卸载
```
make uninstall
```
