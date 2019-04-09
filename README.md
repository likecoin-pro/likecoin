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

## Node REST API
``` 
http://127.0.0.1:8888/<command>? [&pretty] &<param>=<value>.... 
```

##### Get general node and blockchain information
``` 
GET /info 
```

##### Get block 
``` 
GET /block/<blockNum> 
```

##### Get blocks
``` 
GET /blocks?offset=<blockNum>&limit=<countBlocks> 
```

##### Get transaction 
``` 
GET /tx/<txID:hex> 
```

##### Get address info 
``` 
GET /address/<address> 
GET /address/@<username>
GET /address/0x<hexUserID> 
GET /address/?address=<address> 
```

##### Generate new address with Memo 
``` 
GET /address/?address&memo  
```

##### Get address info + memo code 
``` 
GET /address/<address>  
    params:
        [memo=<memo>]
```     
 

##### Get transaction list by address (+memo)
``` 
GET /txs/<address|@username>/
    params: 
        [memo=<num|hex>] 
        [limit=<int>] 
        [order="asc"|"desc"] 
        [offset=<hex>]
```

##### Register new user in blockchain
``` 
POST /new-user?
    params:
        login=<nickname>
        password=<password>
```

##### Generate new key pair, address by secret-phrase
``` 
POST /new-key?
    params: 
        seed=<secret_phrase>
```

##### Transfer founds to address
``` 
POST /new-transfer?
    params: 
        (seed=<secret_phrase>|login&password|private=<hex>) 
        address=<address> 
        [memo=<num|hex>] 
        amount=<integer_in_nano_coins> 
        [comment] 
        [nonce=<num|hex>] 
```

