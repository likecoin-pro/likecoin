# likecoin

## Install likecd node 

##### Install Golang (Linux). version >= 1.11
``` shell
apt-get install golang
```
or
``` shell
wget https://dl.google.com/go/go1.12.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.12.linux-amd64.tar.gz
```

##### Install or update likecd 
``` shell
go get -u -v github.com/likecoin-pro/likecoin/cmd/likecd
go build -o likecd github.com/likecoin-pro/likecoin/cmd/likecd
``` 

##### Show node version, arguments
``` shell
./likecd -version
./likecd -help
```

##### Start likecd node
``` shell
nohup ./likecd -http=localhost:8888 -db=$HOME/likecd.db < /dev/null >/var/log/likecd.log 2>&1 &
``` 

##### Check REST-API
``` shell
http://localhost:8888/info?pretty
```
