package main

import (
	"container/list"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Zilliqa/gozilliqa-sdk/core"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/Zilliqa/gozilliqa-sdk/util"
	"github.com/Zilliqa/gozilliqa-sdk/verifier"
	"io/ioutil"
	"strconv"
	"time"
)

type TxBlockAndDsComm struct {
	TxBlock *core.TxBlock
	DsBlock *core.DsBlock
	DsComm  []core.PairOfNode
}

func printDsComm(dsComm *list.List) {
	cursor := dsComm.Front()
	for cursor != nil {
		fmt.Println(cursor.Value)
		cursor = cursor.Next()
	}
}


func main() {
	//bs, _ := ioutil.ReadFile("1461645.txt")
	//initTxBLock := 1461645
	//initDsBlock := 14617

	bs, _ := ioutil.ReadFile("1454937.txt")
	initTxBLock := 1454937
	initDsBlock := 14550
	blockRawInfo, _ := hex.DecodeString(string(bs))
	var txBlockAndDsComm TxBlockAndDsComm
	_ = json.Unmarshal(blockRawInfo, &txBlockAndDsComm)

	p := provider.NewProvider("https://api.zilliqa.com")

	v := verifier.Verifier{
		NumOfDsGuard: 420,
	}

	initDsComm := list.New()
	for _, ds := range txBlockAndDsComm.DsComm {
		initDsComm.PushBack(ds)
	}

	for {
		currentTxBlock := initTxBLock
		currentTxBlockStr := strconv.FormatUint(uint64(currentTxBlock), 10)
		tst, _ := p.GetTxBlockVerbose(currentTxBlockStr)
		txBlock := core.NewTxBlockFromTxBlockT(tst)
		if txBlock.BlockHeader.DSBlockNum == 18446744073709551615 {
			fmt.Println("wait for a while")
			time.Sleep(time.Second * 30)
			continue
		}
		if txBlock.BlockHeader.DSBlockNum > uint64(initDsBlock) {
			// handle new ds block
			fmt.Println("start validate ds block: ", txBlock.BlockHeader.DSBlockNum)
			dsBlockStr := strconv.FormatUint(txBlock.BlockHeader.DSBlockNum, 10)
			dst, _ := p.GetDsBlockVerbose(dsBlockStr)
			dsBlock := core.NewDsBlockFromDsBlockT(dst)
			newDsComm, err := v.VerifyDsBlock(dsBlock, initDsComm)
			if err != nil {
				panic(err.Error())
			}
			initDsComm = newDsComm
			printDsComm(initDsComm)
			initDsBlock = int(txBlock.BlockHeader.DSBlockNum)
		}
		fmt.Println("start validate tx block: ", currentTxBlock)
		fmt.Println(util.EncodeHex(txBlock.Hash()))
		err := v.VerifyTxBlock(txBlock, initDsComm)
		if err != nil {
			panic(err)
		}

		initTxBLock++
	}

}
