# server backend go

a server backend using go, postgresql, & redis
using domain-driven & test-driven development

__*prequiste:*__

- go: 1.25.0 or higher

<br>

- [__required packages__](./go.mod)

---

## important

__*before anything (build, run, test):*__

1. check on the this root project for config.json:
    - if doesn't exists, copy paste from config.json.template to config.json and modify it as you need

2. check:
    - [backend_api listener](./config.json.template:4)
    - [postgresql main db](./config.json.template:11)
    - [redis main db](./config.json.template:21)

3. scripts:
    - [to build](./run-build.sh)
    - [to test](./run-test.sh)
    - [to debug use dlv](./run-debug.sh)
    - [to run the development backend_api](./run-backend_api.sh)

<br>

__*to check kafka result from websocket:*__

1. run your kafka instance
2. check [producer_ctl](./cmd/producer_ctl/main.go) and adjust your host ipaddrs:port
3. run the producer_ctl on seperate terminal session
4. run the backend_api on seperate terminal session
5. use [websocket wscat](https://github.com/websockets/wscat) and open `/ws/stock/trade`, i.e.:
    - `wscat -c http://hostname:portnumber/ws/stock/trade`
6. consumer has no interaction, it always consume the data as long as it connect to the server

---

<br>

## preview

[quick preview](https://youtu.be/TvGujQngAJ0)

[![](https://img.youtube.com/vi/TvGujQngAJ0/hqdefault.jpg)](https://www.youtube.com/watch?v=TvGujQngAJ0)

go kafka websocket
![go kafka websocket](./docs/img/go-kafka-websocket.png)

<br>

---

###### end of readme

