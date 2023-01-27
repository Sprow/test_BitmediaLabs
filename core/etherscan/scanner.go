package etherscan

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"strconv"
	"test_BitmediaLabs/core/settings"
	"test_BitmediaLabs/core/transactions"
	"test_BitmediaLabs/core/utils"
	"time"
)

const scanLastXBlocks = 1000

type Scanner struct {
	baseURL                string
	reqDelay               time.Duration
	key                    string
	storage                *transactions.MongoStorage
	lastScannedBlockNum    int64
	lastScannedBlockConf   int64
	lastBlockchainBlockNum int64
}

func NewScanner(conf settings.EtherscanAPIConfig, storage *transactions.MongoStorage) *Scanner {
	return &Scanner{
		baseURL:         conf.URL,
		reqDelay:        conf.ReqDelay,
		key:             conf.Key,
		storage:         storage,
	}
}
func (s *Scanner) Start(ctx context.Context) {
	go s.startScanner(ctx)
}

func (s *Scanner) startScanner(ctx context.Context) {
	loop: for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(s.reqDelay):
			if s.lastScannedBlockNum == 0 || s.lastBlockchainBlockNum == s.lastScannedBlockNum {
				s.updateState(ctx)
				log.Println(fmt.Sprintf(
					"state sucsessfuly updated lastBlockInDB=%d lastBlockInBlockchain=%d",
					s.lastScannedBlockNum, s.lastBlockchainBlockNum))
				continue
			}

			nextBlock := s.lastScannedBlockNum + 1
			block, err := s.getBlockByNumber(ctx, nextBlock)
			if err != nil {
				log.Println(errors.Wrap(err, fmt.Sprintf("failed to get block %d", nextBlock)))
				continue
			}
			var TXs []interface{}
			confirmations, err := s.calcConfirmations(ctx)
			if err != nil {
				log.Println(err)
				continue
			}
			for _, txResp := range block.Transactions {
				timestamp, err := utils.HexTimeToTime(block.Timestamp)
				if err != nil {
					log.Println(errors.Wrap(err,
						fmt.Sprintf("failed to conver block timestamp %d", nextBlock)))
					continue loop
				}
				data, err := txResp.ConvertToTX(timestamp, confirmations)
				if err != nil {
					log.Println(errors.Wrap(err,
						fmt.Sprintf("failed to conver txResp to TX  block=%d, txHash=%s", nextBlock, txResp.Hash)))
					continue loop
				}
				TXs = append(TXs, data)
			}
			err = s.storage.BulkInsertTXs(ctx, TXs)
			if err != nil {
				log.Println(errors.Wrap(err, "failed to insert TXs"))
				continue
			}
			log.Println(fmt.Sprintf("block #%d scanned sucssesfuly", nextBlock))
			s.lastScannedBlockNum++
		}
	}
}

// getBlockByNumber return info about block and all TXs
func (s *Scanner) getBlockByNumber(ctx context.Context, blockNum int64) (block Block, err error) {
	blockNumHex := strconv.FormatInt(blockNum, 16)
	url := fmt.Sprintf(
		s.baseURL+
			"?module=proxy"+
			"&action=eth_getBlockByNumber"+
			"&tag=%s"+ // tag = hexBlockNumber with prefix '0x'
			"&boolean=true"+
			"&apikey=%s",
		utils.AddPrefix(blockNumHex), s.key)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return block, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return block, errors.Wrap(err, fmt.Sprintf("failed to get block blockNum=%d", blockNum))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return block, fmt.Errorf("failed to get block blockNum=%d, statuscode=%d", blockNum, resp.StatusCode)
	}

	var response blockResp
	err = json.NewDecoder(resp.Body).Decode(&response)

	return response.Block, nil
}

func (s *Scanner) getLastBlockchainBlockNum(ctx context.Context) error {
	url := fmt.Sprintf(s.baseURL+"?module=proxy&action=eth_getBlockByNumber&tag=latest&boolean=true&apikey=%s", s.key)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to get last block")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get last block, statuscode=%d", resp.StatusCode)
	}

	type blockResp struct {
		Result struct {
			Number string `json:"number"`
		} `json:"result"`
	}

	var block blockResp
	err = json.NewDecoder(resp.Body).Decode(&block)
	if err != nil {
		return err
	}

	s.lastBlockchainBlockNum, err = utils.HexDecimalToDecimal(block.Result.Number)
	return nil
}

func (s *Scanner) updateState(ctx context.Context) {
	err := s.getLastBlockchainBlockNum(ctx)
	if err != nil {
		log.Println(errors.Wrap(err, "failed to get last block number from blockchain"))
		return
	}

	if s.lastScannedBlockNum == 0 {
		blockNum, conf, err := s.storage.FindLastBlockNumAndConfirm(ctx)
		if err != nil {
			log.Println(errors.Wrap(err, "failed to get last block number from db"))
			return
		}
		if blockNum == 0 {
			s.lastScannedBlockNum = s.lastBlockchainBlockNum - scanLastXBlocks
			s.lastScannedBlockConf = scanLastXBlocks
			return
		}
		s.lastScannedBlockNum = blockNum
		s.lastScannedBlockConf = conf
	}
}

func (s *Scanner) calcConfirmations(ctx context.Context) (int64, error) {
	if s.lastScannedBlockConf == 0 {
		err := s.storage.IncAllTXsConf(ctx)
		if err != nil {
			return 0, errors.Wrap(err, "failed to inc all TXs confirmations")
		}
		return 0, nil
	}
	s.lastScannedBlockConf--
	return s.lastScannedBlockConf, nil
}