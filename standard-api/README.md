# standard-api
Go言語による簡単なAPIのサンプル。MySQLのドライバー以外は標準ライブラリにしか依存しません。

## セットアップ
- [Docker & docker-compose](https://docs.docker.com/docker-for-mac/install/)
- [Go 1.11+](https://golang.org/doc/install)

## 実行
- MySQLを起動する
```
docker-compose up
```
- APIサーバを実行する
```
go run main.go
```
-   [Go Modules](https://github.com/golang/go/wiki/Modules)の`GO111MODULE`環境変数を設定していない場合
```
GO111MODULE=on go run main.go 
```

## エンドポイント

`POST /model/create`
```
curl -i -X "POST" -H "Content-Type: application/json" -d '{"name":"My Name"}' "http://localhost:8080/model/create"
```

`DELETE /model/destroy`
```
curl -i -X "DELETE" "http://localhost:8080/model/destroy?id={int}"
```

`PUT /model/edit`
```
curl -i -X "PUT" -H "Content-Type: application/json" -d '{"id":1,"name":"Updated Name"}' "http://localhost:8080/model/edit"
```

`GET /model/index`
```
curl -i -X "GET" "http://localhost:8080/model/index?page={int}&perPage={int}"
```

`GET /model/show/:id`
```
curl -i -X "GET" "http://localhost:8080/model/show?id={int}"
```
