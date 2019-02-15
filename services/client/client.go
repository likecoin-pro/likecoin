package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain"
)

type Client struct {
	apiAddr string
}

func NewClient(apiAddr string) *Client {
	return &Client{
		apiAddr: apiAddr,
	}
}

func (c *Client) httpGet(path string, q url.Values, v interface{}, fn func()) (err error) {
	if q == nil {
		q = url.Values{}
	}
	q.Set("encoding", "binary")

	sURL := c.apiAddr + path + "?" + q.Encode()

	resp, err := http.Get(sURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return
	}
	for r := bin.NewReader(resp.Body); err == nil; {
		if err = r.ReadVar(v); err == nil && fn != nil {
			fn()
		}
	}
	if err == io.EOF {
		return nil
	}
	return
}

func (c *Client) httpPost(path string, data []byte) (err error) {
	sURL := c.apiAddr + path

	resp, err := http.Post(sURL, "binary", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("client.Post(%s)-Error: %v", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("client.Post(%s)-Error: invalid response status code %d", path, resp.StatusCode)
	}
	return
}

func (c *Client) httpGetVal(path string, q url.Values, v interface{}) (err error) {
	return c.httpGet(path, q, v, nil)
}

func (c *Client) GetBlock(num uint64) (block *blockchain.Block, err error) {
	err = c.httpGetVal(fmt.Sprintf("/block/%d", num), nil, &block)
	return
}

func (c *Client) GetBlocks(offset uint64, limit int) (blocks []*blockchain.Block, err error) {
	var block *blockchain.Block
	err = c.httpGet("/blocks", url.Values{
		"offset": {fmt.Sprint(offset)},
		"limit":  {fmt.Sprint(limit)},
	}, &block, func() {
		if block != nil {
			blocks = append(blocks, block)
		}
	})
	return
}

func (c *Client) PutTx(tx *blockchain.Transaction) (err error) {
	return c.PutTxs([]*blockchain.Transaction{tx})
}

func (c *Client) PutTxs(txs []*blockchain.Transaction) (err error) {
	return c.httpPost("/new-txs", bin.Encode(txs))
}
