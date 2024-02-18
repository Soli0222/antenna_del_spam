# Antenna Delete SPAM

Misskeyにおける、アンテナを用いたスパム撃退ツールです。  

アンテナの中に入ったメンションが2つ以上ついているノートを対象とし、そのノート,ユーザーを削除し、ドメインブロックを行います。

## 利用方法

1. Misskeyで自ドメインのアンテナを作成する
2. `git clone https://github.com/Soli0222/antenna_del_spam`でリポジトリをクローン
3. `cp .env.example .env`をして`.env`ファイルにドメイン、トークン、アンテナIDを設定
4. `go run main.go`