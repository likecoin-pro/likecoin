package client

import (
	"fmt"
	"io/ioutil"
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

func (c *Client) httpRequest(path string, q url.Values, v interface{}) (err error) {
	if q == nil {
		q = url.Values{}
	}
	q.Set("encoding", "binary")

	sURL := c.apiAddr + path + "?" + q.Encode()

	resp, err := http.Get(sURL)
	//log.Print("== http request ", sURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	//log.Print("== RESP code:", resp.StatusCode, " len:", len(data))
	if err != nil {
		return err
	}
	err = bin.Decode(data, v)
	return
}

func (c *Client) GetBlock(num uint64) (block *blockchain.Block, err error) {
	err = c.httpRequest(fmt.Sprintf("/block/%d", num), nil, &block)
	return
}
