package etherscan

import (
	"fmt"
	"github.com/pkg/errors"
	"test_BitmediaLabs/core/transactions"
	"test_BitmediaLabs/core/utils"
	"time"
)

type blockResp struct {
	Id    int `json:"id"`
	Block `json:"result"`
}

type Block struct {
	Number       string   `json:"number"`
	Timestamp    string   `json:"timestamp"`
	Transactions []TXResp `json:"transactions"`
}

type TXResp struct {
	Hash        string `json:"hash"`
	From        string `json:"from"`
	To          string `json:"to"`
	BlockNumber string `json:"blockNumber"`
	Gas         string `json:"gas"`
	GasPrice    string `json:"gasPrice"`
	Value       string `json:"value"`
}

func (r *TXResp) ConvertToTX(timestamp time.Time, confirmations int64) (tx transactions.TX, err error) {
	blockNum, err := utils.HexDecimalToDecimal(r.BlockNumber)
	if err != nil {
		return tx, errors.Wrap(err, fmt.Sprintf("failed to insert r data, hash='%s'", utils.CutPrefix(r.Hash)))
	}
	valueETH := r.convertValueToEth()

	feeETH, err := r.calculateFee()
	if err != nil {
		return
	}

	tx = transactions.TX{
		Hash:          utils.CutPrefix(r.Hash),
		From:          utils.CutPrefix(r.From),
		To:            utils.CutPrefix(r.To),
		BlockNumber:   blockNum,
		Value:         valueETH,
		Fee:           feeETH,
		Confirmations: confirmations,
		Timestamp:     timestamp,
	}
	return
}

// calculateFee return TXResp fee unit:Gwei
// Так і не вийшло порахувати fee правильно
// etherscan апі на запит eth_getTransactionByHash вітдає поля "maxFeePerGas" і "maxPriorityFeePerGas"
// Але описання цих полів немає ні на https://docs.etherscan.io/api-endpoints/geth-parity-proxy#eth_gettransactionbyhash
// Ні на https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_gettransactionbyhash
func (r *TXResp) calculateFee() (float64, error) {
	gas, err := utils.HexDecimalToDecimal(r.Gas)
	if err != nil {
		return 0, err
	}
	gasPriceWei := utils.HexDecimalToFloat64(r.GasPrice)
	feeEth := float64(gas) * utils.WeiToEth(gasPriceWei)

	return feeEth, nil
}

func (r *TXResp) convertValueToEth() float64 {
	valueWei := utils.HexDecimalToFloat64(r.Value)
	return utils.WeiToEth(valueWei)
}
