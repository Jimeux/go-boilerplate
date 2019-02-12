# standard-api
### Go言語による簡単なAPIサーバのサンプルです
[MySQLのドライバー](https://github.com/go-sql-driver/mysql)以外は**標準ライブラリにしか依存しません**。より簡潔なAPIを書くために以下のライブラリを検討してください。

- ルーター／フレームワーク：[mux](https://github.com/gorilla/mux)、[Gin](https://github.com/gin-gonic/gin)
- ORM：[Xorm](https://github.com/go-xorm/xorm)、[Gorm](https://github.com/jinzhu/gorm)

## セットアップ
以下をインストールする。
- [Docker & docker-compose](https://docs.docker.com/docker-for-mac/install/)
- [Go 1.11](https://golang.org/doc/install)

## 実行
- MySQLを起動する。初回は`db/init/init.sql`の中身がDBに入れられる。
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

## APIのエンドポイント

`POST /model/create`
```
curl -i -X "POST" -H "Content-Type: application/json" -d '{"name":"My Name"}' "http://localhost:8080/model/create"
```

`DELETE /model/destroy`
```
curl -i -X "DELETE" "http://localhost:8080/model/destroy?id=1"
```

`PUT /model/edit`
```
curl -i -X "PUT" -H "Content-Type: application/json" -d '{"id":1,"name":"Updated Name"}' "http://localhost:8080/model/edit"
```

`GET /model/index`
```
curl -i -X "GET" "http://localhost:8080/model/index?page=1&perPage=5"
```

`GET /model/show/:id`
```
curl -i -X "GET" "http://localhost:8080/model/show?id=2"
```

＊ スラッシュの有無にご注意ください。

## ファイルの説明

### `app/model.go`
`model`テーブルの構造体（`Model`）を定義する

### `app/dao.go`
`model`テーブルのCRUDを行う構造体（`DAO`）を定義する。[`sql.DB`](https://golang.org/pkg/database/sql/#DB)に依存する。

### `app/controller.go`
HTTPリクエスト／レスポンスとロギングに対応する構造体（`Controller`）を定義する。`DAO`に依存する。

### `main.go`
DBなどの初期化を行い、エンドポイントを定義しHTTPサーバを起動する。

### `go.mod`/`go.sum`
Goモジュールのコマンドに自動で生成・更新されるファイル。手動で変更しない。
