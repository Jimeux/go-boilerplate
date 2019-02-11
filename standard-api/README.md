# standard-api
Go言語による簡単なAPIのサンプル。MySQLのドライバー以外は標準ライブラリにしか依存しません。

- 

## セットアップ
- [Docker & docker-compose](https://docs.docker.com/docker-for-mac/install/)
- Go 1.11+

## 実行
```
docker-compose up
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

`GET /model/index`
```
curl -i "http://localhost:8080/model/index?page={int}&perPage={int}"
```

`GET /model/show/:id`
```
curl -i "http://localhost:8080/model/show?id={int}"
```
