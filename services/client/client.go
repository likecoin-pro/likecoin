package client

import (
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

func (c *Client) httpRequest(path string, q url.Values, v interface{}, fn func()) (err error) {
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
		if err = r.ReadVar(v); err == nil {
			fn()
		}
	}
	if err == io.EOF {
		return nil
	}
	return
}

func (c *Client) httpRequestVal(path string, q url.Values, v interface{}) (err error) {
	return c.httpRequest(path, q, v, func() {})
}

func (c *Client) GetBlock(num uint64) (block *blockchain.Block, err error) {
	err = c.httpRequestVal(fmt.Sprintf("/block/%d", num), nil, &block)
	return
}

func (c *Client) GetBlocks(offset uint64, limit int) (blocks []*blockchain.Block, err error) {
	var block *blockchain.Block
	err = c.httpRequest("/blocks", url.Values{
		"offset": {fmt.Sprint(offset)},
		"limit":  {fmt.Sprint(limit)},
	}, &block, func() {
		blocks = append(blocks, block)
	})
	return
}
