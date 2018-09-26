package webapi

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/db"
	"github.com/likecoin-pro/likecoin/commons/bignum"
	hex2 "github.com/likecoin-pro/likecoin/commons/hex"
	"github.com/likecoin-pro/likecoin/commons/log"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/likecoin-pro/likecoin/object"
)

type Context struct {
	req       *http.Request
	rw        http.ResponseWriter
	bc        *db.BlockchainStorage
	urlPrefix string
}

func NewContext(
	req *http.Request,
	rw http.ResponseWriter,
	bc *db.BlockchainStorage,
	urlPrefix string,
) *Context {
	return &Context{
		req:       req,
		rw:        rw,
		bc:        bc,
		urlPrefix: urlPrefix,
	}
}

var (
	reBlockNum   = regexp.MustCompile(`^/block/(\d{1,12})$`)             //
	reBlockTxNum = regexp.MustCompile(`^/block/(\d{1,12})/(\d{1,12})$`)  //
	reTxID       = regexp.MustCompile(`^/tx/([a-f0-9]{1,16})$`)          //
	reTxHash     = regexp.MustCompile(`^/tx/([a-f0-9]{64})$`)            //
	reAddrInfo   = regexp.MustCompile(`^/address/(@?[0-9a-zA-Z]+)$`)     //
	reUserInfo   = regexp.MustCompile(`^/user/(@?[0-9a-zA-Z]+)$`)        //
	reTxsAddr    = regexp.MustCompile(`^/txs/(@?[0-9a-zA-Z]+)$`)         //
	reAddrTxs    = regexp.MustCompile(`^/address/(@?[0-9a-zA-Z]+)/txs$`) //
)

/**

API /v0/

./ 								-> html dashboard

./blocks						-> [{block},...]
	&offset=<blockNum:int>
	&limit=<limit:int>
	&order=asc|desc

./block/<num:int|hash:hex>		-> {block}

./tx/<txID|txHash:hex>			-> {tx}

./txs   						-> [{tx},...]
	&address=<address>
	&asset=<asset:hex>
	&memo=<memo:uint64>
	&offset=<ts:int>
	&limit=<limit:int>
	&order=asc|desc

./txs/<address>					-> synonym of /txs/?address=<address>

./user/<address>				-> {userTx}
./user/@<username>				-> {userTx}

./address/<address>				-> {addressInfo, balance}
	&memo
	&asset

<address> := "LikeXXXXXXXXXXXXXX" | <pubKey:base58> | @<nick> | 0x<userID:hex>

*/
func (ctx *Context) Exec() {

	log.Trace.Printf("webapi: HTTP-request: %s  PATH: %s", ctx.req.RequestURI, ctx.req.URL.Path)

	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(*HTTPError); ok && e != nil {
				if e.Code != 404 {
					log.Error.Printf("HTTP-ERROR-%d: %s", e.Code, e.Err)
				}
				http.Error(ctx.rw, e.Err, e.Code)

			} else if err, ok := r.(error); ok && err != nil {
				log.Error.Printf("HTTP-ERROR: %v", err)
				http.Error(ctx.rw, err.Error(), http.StatusInternalServerError)
			}
		}
	}()

	if err := ctx.req.ParseForm(); err != nil {
		ctx.Panic400(err)
	}

	path := ctx.req.URL.Path
	if !strings.HasPrefix(path, ctx.urlPrefix) {
		ctx.Panic400Str("invalid path")
	}
	path = strings.TrimPrefix(path, ctx.urlPrefix)
	if len(path) > 1 {
		path = strings.TrimRight(path, "/")
	}

	var q []string
	pathMatch := func(re *regexp.Regexp) bool {
		q = re.FindStringSubmatch(path)
		return len(q) > 0
	}

	switch {

	// /  -> dashboard html-page
	//case path == "/":
	//	ctx.WriteHTML(htmlDashboard)
	//	return

	case path == "/info":
		ctx.WriteObject(ctx.bc.Info())

	case path == "/mempool/txs":
		if ctx.Get("address", "") != "" {
			addr, _, _ := ctx.getAddress()
			ctx.WriteObject(ctx.bc.Mempool.TxsByAddress(addr))
		} else {
			ctx.WriteObject(ctx.bc.Mempool.TxsAll())
		}

	case path == "/new-transfer":
		prv := ctx.getPrvKey()                  // prv OR secret
		addr, toMemo, asset := ctx.getAddress() // address
		amount := ctx.getAmount("amount")       // amount
		comment := ctx.Get("comment", "")       // comment

		tx := object.NewSimpleTransfer(ctx.bc.Cfg, prv, addr, amount, asset, comment, 0, toMemo)
		if err := tx.Verify(ctx.bc.Cfg); err != nil {
			ctx.Panic400(err)
		}
		ctx.bc.Mempool.Put(tx)
		ctx.WriteObject(tx)

	case path == "/memo-address":
		addr, memo, _ := ctx.getAddress() // address
		ctx.WriteObject(struct {
			Address     string `json:"address"`
			Memo        string `json:"memo"`
			MemoAddress string `json:"memo_address"`
		}{
			addr.String(),
			"0x" + hex2.EncodeUint(memo),
			addr.MemoString(memo),
		})

	case path == "/new-user":
		prv := ctx.getPrvKey()             // prv OR secret
		nick := ctx.Get("nick", "")        // users nickname
		refID := ctx.getUint("ref", 0, 16) // referrer

		tx := object.NewUser(ctx.bc.Cfg, prv, nick, refID, nil)
		if err := tx.Verify(ctx.bc.Cfg); err != nil {
			ctx.Panic400(err)
		}
		ctx.bc.Mempool.Put(tx)
		ctx.WriteObject(tx)

		// /blocks
	case path == "/blocks":
		ofst, limit, ord := ctx.getOffset(), ctx.getLimit(), ctx.getOrder("asc")
		strm := ctx.OpenStream()
		defer strm.Close()
		ctx.bc.FetchBlocks(ofst, limit, ord, func(block *blockchain.Block) error {
			return strm.WriteObject(block)
		})

	// 	/txs/<address>  OR   /address/<address>/txs
	case pathMatch(reTxsAddr) || pathMatch(reAddrTxs):
		addr, memo, asset, offset, limit, order := ctx.parseQueryParams(q[1])
		strm := ctx.OpenStream()
		defer strm.Close()
		ctx.bc.FetchTransactionsByAddr(asset, addr, memo, offset, limit, order, func(tx *blockchain.Transaction, _ bignum.Int) error {
			return strm.WriteObject(tx)
		})

		// 	/block/<num>
	case pathMatch(reBlockNum):
		num, _ := strconv.ParseUint(q[1], 0, 64)
		if block, err := ctx.bc.GetBlock(num); err == db.ErrBlockNotFound {
			ctx.Panic404(err)
		} else {
			ctx.WriteObject(block, err)
		}

		// 	/block/<num>/<num>
	case pathMatch(reBlockTxNum):
		blockNum, _ := strconv.ParseUint(q[1], 0, 64)
		txIdx, _ := strconv.ParseUint(q[2], 0, 64)
		ctx.WriteObject(ctx.bc.GetTransaction(blockNum, int(txIdx)))

		// 	/user/<userID:hex|@nick|address|pubkey>
	case pathMatch(reUserInfo):
		userID := ctx.parseUserID(q[1])
		tx, _, err := ctx.bc.UserByID(userID)
		ctx.WriteObject(tx, err)

		//	/tx/<hash:hex>
	case pathMatch(reTxHash):
		txHash, _ := hex.DecodeString(q[1])
		ctx.WriteObject(ctx.bc.TransactionByHash(txHash))

		//	/tx/<txID:hex>
	case pathMatch(reTxID):
		id, _ := strconv.ParseUint(q[1], 16, 64)
		ctx.WriteObject(ctx.bc.TransactionByID(id))

		// 	/address?address&memo&asset
	case path == "/address":
		ctx.WriteObject(ctx.bc.AddressInfo(ctx.getAddress()))

		// 	/address/<address>?memo&asset
	case pathMatch(reAddrInfo):
		ctx.WriteObject(ctx.bc.AddressInfo(ctx.parseAddress(q[1])))

	default:
		ctx.Panic404(err404)
	}
}

var err404 = errors.New("not found")

type HTTPError struct {
	Code int
	Err  string
}

func (c *Context) Panic(code int, err error) {
	panic(&HTTPError{code, err.Error()})
}

func (c *Context) Panic400(err error) {
	c.Panic(http.StatusBadRequest, err)
}

func (c *Context) Panic404(err error) {
	c.Panic(http.StatusNotFound, err)
}

func (c *Context) Panic500(err error) {
	c.Panic(http.StatusInternalServerError, err)
}

// deprecated
func (c *Context) Panic400Str(err string) {
	c.Panic400(errors.New(err))
}

//func (c *Context) SetMaxExpire() {
//	c.SetExpire(time.Hour * 24 * 365 * 10)
//}
//
//func (c *Context) SetExpire(t time.Duration) {
//	c.rw.Header().Set("Cache-Control", fmt.Sprintf("max-age:%d, public", int(t.Seconds())))
//	c.rw.Header().Set("Last-Modified", time.Now().Format(http.TimeFormat))
//	c.rw.Header().Set("Expires", time.Now().Add(t).Format(http.TimeFormat))
//}

func (c *Context) WriteHTML(data []byte, ee ...error) {
	if len(ee) > 0 && ee[0] != nil {
		c.Panic500(ee[0])
		return
	}
	c.rw.Header().Add("Content-Type", "text/html; charset=utf-8")
	c.rw.WriteHeader(http.StatusOK)
	if _, err := io.Copy(c.rw, bytes.NewBuffer(data)); err != nil {
		log.Error.Printf("HTTP-Response-ERROR: %v", err)
	}
}

func (c *Context) marshalJSON(v interface{}) (data []byte, err error) {
	if _, prettyJSON := c.req.Form["pretty"]; prettyJSON {
		data, err = json.MarshalIndent(v, "", "  ")
	} else {
		data, err = json.Marshal(v)
	}
	return
}

func (c *Context) WriteObject(obj interface{}, ee ...error) {
	if len(ee) > 0 && ee[0] != nil {
		c.Panic500(ee[0])
		return
	}

	var data []byte
	var err error
	switch c.Get("encoding", "json") {
	case "json":
		c.rw.Header().Set("Content-Type", "text/json; charset=utf-8")
		data, err = c.marshalJSON(obj)

	case "binary":
		c.rw.Header().Set("Content-Type", "binary")
		data = bin.Encode(obj)
	}
	if err != nil {
		c.Panic500(err)
	}
	c.rw.WriteHeader(http.StatusOK)
	if _, err := io.Copy(c.rw, bytes.NewBuffer(data)); err != nil {
		log.Error.Printf("HTTP-Response-ERROR: %v", err)
	}
}

func (c *Context) OpenStream() (s *RWStream) {
	s = &RWStream{
		rw:     c.rw,
		pretty: len(c.req.Form["pretty"]) > 0,
	}
	var err error
	switch c.Get("encoding", "") {
	case "json", "":
		s.encoding = encodingJSON
		c.rw.Header().Set("Content-Type", "text/json; charset=utf-8")
		c.rw.WriteHeader(http.StatusOK)
		_, err = io.Copy(c.rw, bytes.NewBufferString("[\n"))
	case "binary":
		s.encoding = encodingBinary
		c.rw.Header().Set("Content-Type", "binary")
		c.rw.WriteHeader(http.StatusOK)
	default:
		c.Panic400Str("incorrect encoding")
	}
	if err != nil {
		c.Panic500(err)
	}
	return
}

type RWStream struct {
	rw       http.ResponseWriter
	encoding int
	pretty   bool
	cnt      int64
}

const (
	encodingJSON   = 0
	encodingBinary = 1
)

func (s *RWStream) WriteObject(v interface{}) (err error) {
	var buf = bytes.NewBuffer(nil)
	switch s.encoding {
	case encodingJSON:
		if s.cnt > 0 {
			buf.Write([]byte(",\n"))
		}
		var data []byte
		if s.pretty {
			data, err = json.MarshalIndent(v, "", "  ")
		} else {
			data, err = json.Marshal(v)
		}
		buf.Write(data)
	case encodingBinary:
		buf.Write(bin.Encode(v))
	}
	if err == nil {
		_, err = io.Copy(s.rw, buf)
	}
	s.cnt++
	return
}

func (s *RWStream) Close() (err error) {
	var buf = bytes.NewBuffer(nil)
	switch s.encoding {
	case encodingJSON:
		buf.WriteString("\n]")
	case encodingBinary:
		buf.WriteByte(0)
	}
	_, err = io.Copy(s.rw, buf)
	return
}

/*

	c.rw.Header().Add("Content-Type", "text/json; charset=utf-8")
	c.rw.WriteHeader(http.StatusOK)

	_, err := c.rw.Write([]byte("["))
	if err != nil {
		log.Error.Printf("HTTP-Response-ERROR: %v", err)
		return
	}

	first := true
	err = fn(func(obj interface{}) (err error) {

		data, err := c.marshalJSON(obj)
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(nil)
		if first {
			first = false
		} else {
			buf.WriteString(",\n")
		}
		buf.Write(data)

		_, err = io.Copy(c.rw, buf)
		return
	})
	if err != nil {
		log.Error.Printf("HTTP-Response-ERROR: %v", err)
		return
	}

	_, err = c.rw.Write([]byte("]"))
	if err != nil {
		log.Error.Printf("HTTP-Response-ERROR: %v", err)
		return
	}
}
*/

//--------------- params -------------------
func (c *Context) Get(name, defaultVal string) string {
	if s := strings.TrimSpace(c.req.Form.Get(name)); s != "" {
		return s
	}
	return defaultVal
}
func (c *Context) getUint(name string, defaultVal uint64, base int) uint64 {
	if s := strings.TrimSpace(c.req.Form.Get(name)); s != "" {
		i, err := strconv.ParseUint(s, base, 64)
		if err != nil {
			c.Panic400Str("incorrect param " + name)
		}
		return i
	}
	return defaultVal
}

func (c *Context) getAddress() (addr crypto.Address, memo uint64, asset assets.Asset) {
	return c.parseAddress(c.Get("address", ""))
}

func (c *Context) getMemo() uint64 {
	return c.getUint("memo", 0, 0)
}

func (c *Context) getPrvKey() *crypto.PrivateKey {
	prv, err := crypto.ParsePrivateKey(c.Get("prv", ""))
	if err != nil {
		c.Panic400Str("incorrect prv-param")
	}
	return prv
}

func (c *Context) getAmount(param string) bignum.Int {
	v, err := strconv.ParseUint(c.Get(param, "0"), 10, 64)
	if err != nil {
		c.Panic400Str("incorrect amount-param")
	}
	return bignum.NewInt(int64(v))
}

var defaultAsset = assets.Default.String()

func (c *Context) getAsset() assets.Asset {
	asset, err := assets.ParseAsset(c.Get("asset", defaultAsset))
	if err != nil {
		c.Panic400Str("incorrect asset-params")
	}
	return asset
}

func (c *Context) getLimit() int64 {
	n, err := strconv.ParseInt(c.Get("limit", "100"), 0, 64)
	if n <= 0 || err != nil {
		c.Panic400Str("incorrect limit-param")
	}
	if n > 10000 {
		c.Panic400Str("limit-param is too much")
	}
	return n
}

func (c *Context) getOffset() uint64 {
	n, err := strconv.ParseUint(c.Get("offset", "0"), 0, 64)
	if n < 0 || err != nil {
		c.Panic400Str("incorrect offset-params")
	}
	return n
}

func (c *Context) getOrder(defaultValue string) (desc bool) {
	switch strings.ToLower(c.Get("order", defaultValue)) {
	case "asc", "":
		return false
	case "desc":
		return true
	}
	c.Panic400Str("incorrect order-params")
	return
}

func (c *Context) parseQueryParams(sAddr string) (addr crypto.Address, memo uint64, asset assets.Asset, offset uint64, limit int64, order bool) {
	if sAddr != "" {
		addr, memo, asset = c.parseAddress(sAddr)
	}
	offset = c.getOffset()
	limit = c.getLimit()
	order = c.getOrder("desc")
	return
}

func (c *Context) parseAddress(s string) (addr crypto.Address, memo uint64, asset assets.Asset) {
	if s == "" {
		s = c.Get("address", "")
	}
	addr, memo, err := c.bc.AddressByStr(s)
	if err != nil {
		if strings.HasSuffix(err.Error(), "not found") {
			c.Panic404(err)
		}
		c.Panic400Str("incorrect address-params.\nError: " + err.Error())
	}
	asset = c.getAsset()
	if m := c.getMemo(); m != 0 {
		memo = m
	}
	return
}

func (c *Context) parseUserID(s string) (userID uint64) {
	if strings.HasPrefix(s, "@") { // @<nick>
		if addr, _, err := c.bc.NameAddress(s[1:]); err != nil {
			panic(err)
		} else {
			return addr.ID()
		}
	} else if strings.HasPrefix(s, "0x") { // 0x<userID:hex>
		if userID, err := strconv.ParseUint(s, 0, 64); err != nil {
			c.Panic400Str("incorrect address-param")
		} else {
			return userID
		}
	}
	addr, _, err := crypto.ParseAddress(s)
	if err != nil || addr.Empty() {
		c.Panic400Str("incorrect address-params")
	}
	return addr.ID()
}
