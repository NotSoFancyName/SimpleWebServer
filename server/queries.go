package server

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
)

const (
	queryURL = "https://api.etherscan.io/api?module=proxy&boolean=true"

	getBlockAction       = "eth_getBlockByNumber"
	getTransactionsCount = "eth_getBlockTransactionCountByNumber"

	apiKey = "RZ5PGXXEU6FRZZTWJH7QKPSKMRZRKPFWAN"
)

func queryBlockInfo(blockNum int) (*blockInfo, error) {
	u, err := url.Parse(queryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}
	q := u.Query()
	q.Set("apikey", apiKey)
	q.Set("tag", fmt.Sprintf("%x", blockNum))
	q.Set("action", getBlockAction)
	u.RawQuery = q.Encode()
	client := &http.Client{}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query 'api.etherscan.io': %v", err)
	}
	defer resp.Body.Close()

	var etherumBlock struct {
		Result struct {
			Transactions []struct {
				Value string `json:"value"`
			} `json:"transactions"`
		} `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&etherumBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body': %v", err)
	}

	amount := &big.Int{}
	for _, v := range etherumBlock.Result.Transactions {
		transactionAmount, ok := new(big.Int).SetString(v.Value, 0)
		if !ok {
			return nil, fmt.Errorf("failed to parse transaction value: %v", err)
		}
		amount.Add(amount, transactionAmount)
	}

	etherumAmount := new(big.Float).SetInt(amount)
	e, _ := etherumAmount.Mul(etherumAmount, wieToEthRatio).Float64()

	bi := &blockInfo{
		Amount:       e,
		Transactions: len(etherumBlock.Result.Transactions),
	}

	return bi, nil
}

func queryTransactionsCount(blockNum int) (int, error) {
	u, err := url.Parse(queryURL)
	if err != nil {
		return -1, fmt.Errorf("failed to parse URL: %v", err)
	}
	q := u.Query()
	q.Set("apikey", apiKey)
	q.Set("tag", fmt.Sprintf("%x", blockNum))
	q.Set("action", getTransactionsCount)
	u.RawQuery = q.Encode()
	client := &http.Client{}
	resp, err := client.Get(u.String())
	if err != nil {
		return -1, fmt.Errorf("failed to query 'api.etherscan.io': %v", err)
	}
	defer resp.Body.Close()

	var count struct {
		Result string `json:"result"`
	}
	err = json.NewDecoder(resp.Body).Decode(&count)
	if err != nil {
		return -1, fmt.Errorf("failed to decode response body': %v", err)
	}

	i, err := strconv.ParseInt(count.Result, 0, 0)
	if err != nil {
		return -1, fmt.Errorf("failed to parse transactions counter': %v", err)
	}

	return int(i), nil
}
