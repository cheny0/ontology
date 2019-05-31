package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/store/ledgerstore"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/syndtr/goleveldb/leveldb"
	_ "net/http/pprof"
	"os"
	"sync"
	"time"
)

type ExecuteInfo struct {
	Height   uint32
	ReadSet  *overlaydb.MemDB
	WriteSet *overlaydb.MemDB
	GasTable map[string]uint64
}

func main() {
	checkAllBlock()
}

func checkOneBlock() {
	blockHeight := uint32(534300)
	blockHeight = uint32(1294201)
	ledgerstore.MOCKDBSTORE = false

	dbDir := "./Chain/ontology"

	modkDBPath := fmt.Sprintf("%s%s%s", dbDir, string(os.PathSeparator), "states"+"mockdb")
	levelDB, err := ledgerstore.OpenLevelDB(modkDBPath)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	ledgerStore, err := ledgerstore.NewLedgerStore(dbDir, 3000000)
	initLedgerStore(ledgerStore)

	executeInfo, err := getExecuteInfoByHeight(blockHeight, levelDB)
	execute(executeInfo, ledgerStore)
}

func checkAllBlock() {
	var wg = new(sync.WaitGroup)

	ledgerstore.MOCKDBSTORE = false

	dbDir := "./Chain/ontology"

	modkDBPath := fmt.Sprintf("%s%s%s", dbDir, string(os.PathSeparator), "states"+"mockdb")
	levelDB, err := ledgerstore.OpenLevelDB(modkDBPath)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	ledgerStore, err := ledgerstore.NewLedgerStore(dbDir, 3000000)
	initLedgerStore(ledgerStore)

	start := time.Now()

	ch := make(chan interface{}, 100)
	currentBlockHeight := ledgerStore.GetCurrentBlockHeight()

	for i := uint32(0); i < 4; i++ {
		wg.Add(1)
		go sendExecuteInfoToCh(ch, i, currentBlockHeight, levelDB, wg)
	}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go handleExecuteInfo(ch, ledgerStore, wg)
	}

	//log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
	wg.Wait()
	fmt.Println("Current BlockHeight: ", ledgerStore.GetCurrentBlockHeight())
	fmt.Println("start: ", start)
	fmt.Println("end: ", time.Now())
}

func handleExecuteInfo(ch <-chan interface{}, ledgerStore *ledgerstore.LedgerStoreImp, wg *sync.WaitGroup) {
	for {
		select {
		case task, ok := <-ch:
			if !ok {
				wg.Done()
				return
			}
			executeInfo, ok := task.(*ExecuteInfo)
			if ok {
				execute(executeInfo, ledgerStore)
			} else {
				wg.Done()
			}
		}
	}
}

func sendExecuteInfoToCh(ch chan<- interface{}, offset uint32, currentBlockHeight uint32, levelDB *leveldb.DB, wg *sync.WaitGroup) {
	for i := uint32(323510); 4*i+offset < currentBlockHeight; i++ {
		executeInfo, err := getExecuteInfoByHeight(4*i+offset, levelDB)
		if err != nil {
			fmt.Println("err:", err)
			return
		}
		ch <- executeInfo
	}
	ch <- "success"
	wg.Done()
}

func execute(executeInfo *ExecuteInfo, ledgerStore *ledgerstore.LedgerStoreImp) {

	overlay := overlaydb.NewOverlayDB(ledgerstore.NewMockDBWithMemDB(executeInfo.ReadSet))
	hash := ledgerStore.GetBlockHash(executeInfo.Height)
	block, err := ledgerStore.GetBlockByHash(hash)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	refreshGlobalParam(executeInfo.GasTable)
	cache := storage.NewCacheDB(overlay)
	//overlaydb.IS_SHOW = false
	neovm.PrintOpcode = false
	//index := 0
	for _, tx := range block.Transactions {
		cache.Reset()
		//fmt.Fprintf(os.Stderr, "begin transaction, index:%d\n", index)
		_, e := handleTransaction(ledgerStore, overlay, cache, block, tx)
		//fmt.Fprintf(os.Stderr, "end transaction, index:%d\n", index)
		//index++
		if e != nil {
			fmt.Println("err:", e)
			return
		}
	}
	//overlaydb.IS_SHOW = false

	writeSet := overlay.GetWriteSet()
	//fmt.Printf("hash:  %x", writeSet.Hash())
	//fmt.Println("*****************")
	//fmt.Println("*****************")
	//fmt.Printf("hash:  %x", executeInfo.WriteSet.Hash())

	if !bytes.Equal(writeSet.Hash(), executeInfo.WriteSet.Hash()) {

		//writeSet.Hash()
		//fmt.Println("**********************")
		//executeInfo.WriteSet.Hash()

		//tempMap := make(map[string]string)
		//writeSet.ForEach(func(key, val []byte) {
		//	tempMap[common.ToHexString(key)] = common.ToHexString(val)
		//})
		//executeInfo.WriteSet.ForEach(func(key, val []byte) {
		//	if tempMap[common.ToHexString(key)] != common.ToHexString(val) {
		//		fmt.Printf("key:%x, value: %x\n", key, val)
		//	}
		//})

		fmt.Printf("blockheight:%d, writeSet.Hash:%x, executeInfo.WriteSet.Hash:%x\n", executeInfo.Height, writeSet.Hash(), executeInfo.WriteSet.Hash())
		panic(executeInfo.Height)
	}

	fmt.Println("blockHeight: ", executeInfo.Height)

	//fmt.Fprintf(os.Stderr, "diff hash at height:%d, hash:%x\n", block.Header.Height, writeSet.Hash())
	//
	//fmt.Fprintf(os.Stderr, "diff hash at height:%d, hash:%x\n", block.Header.Height, executeInfo.WriteSet.Hash())
}

func getExecuteInfoByHeight(height uint32, levelDB *leveldb.DB) (*ExecuteInfo, error) {
	//get gasTable
	key := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(key[:], height)
	dataBytes, err := levelDB.Get(key, nil)
	if err != nil {
		return nil, fmt.Errorf("get databytes error: %s", err)
	}
	source := common.NewZeroCopySource(dataBytes)
	l, eof := source.NextUint32()
	if eof {
		return nil, fmt.Errorf("gastable length is wrong: %d", l)
	}

	m := make(map[string]uint64)
	for i := uint32(0); i < l; i++ {
		key, _, irregular, eof := source.NextVarBytes()
		if irregular || eof {
			return nil, fmt.Errorf("update gastable NextVarBytes error")
		}
		val, eof := source.NextUint64()
		if eof {
			return nil, fmt.Errorf("update gastable NextUint64 error")
		}
		m[string(key)] = val
	}
	//get readSet
	l, eof = source.NextUint32()
	if eof {
		return nil, fmt.Errorf("readset NextUint32 error: %d", l)
	}
	readSetDB := overlaydb.NewMemDB(16*1024, 16)
	for i := uint32(0); i < l; i++ {
		key, _, irregular, eof := source.NextVarBytes()
		if eof || irregular {
			break
		}
		value, _, _, eof := source.NextVarBytes()
		if eof {
			break
		}
		readSetDB.Put(key, value)
	}

	// get writeSet
	l, eof = source.NextUint32()
	if eof {
		return nil, fmt.Errorf("writeset NextUint32 error: %d", l)
	}
	writeSetDB := overlaydb.NewMemDB(16*1024, 16)
	for i := uint32(0); i < l; i++ {
		key, _, irregular, eof := source.NextVarBytes()
		if eof || irregular {
			break
		}
		value, _, _, eof := source.NextVarBytes()
		if eof {
			break
		}
		if height == 54 {
			log.Errorf("key:%x, value:%x", key, value)
		}
		writeSetDB.Put(key, value)
	}
	return &ExecuteInfo{Height: height, ReadSet: readSetDB, WriteSet: writeSetDB, GasTable: m}, nil
}

func initLedgerStore(ledgerStore *ledgerstore.LedgerStoreImp) {
	bookKeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	genesisConfig := config.DefConfig.Genesis
	genesisBlock, err := genesis.BuildGenesisBlock(bookKeepers, genesisConfig)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	ledgerStore.InitLedgerStoreWithGenesisBlock(genesisBlock, bookKeepers)
}

func handleTransaction(ledgerStore *ledgerstore.LedgerStoreImp, overlay *overlaydb.OverlayDB, cache *storage.CacheDB, block *types.Block, tx *types.Transaction) (*event.ExecuteNotify, error) {
	txHash := tx.Hash()
	notify := &event.ExecuteNotify{TxHash: txHash, State: event.CONTRACT_STATE_FAIL}
	stateStore := ledgerstore.StateStore{}
	switch tx.TxType {
	case types.Deploy:
		err := stateStore.HandleDeployTransaction(ledgerStore, overlay, cache, tx, block, notify)
		if overlay.Error() != nil {
			return nil, fmt.Errorf("HandleDeployTransaction tx %s error %s", txHash.ToHexString(), overlay.Error())
		}
		if err != nil {
			log.Debugf("HandleDeployTransaction tx %s error %s", txHash.ToHexString(), err)
		}
	case types.Invoke:
		err := stateStore.HandleInvokeTransaction(ledgerStore, overlay, cache, tx, block, notify)
		if overlay.Error() != nil {
			return nil, fmt.Errorf("HandleInvokeTransaction tx %s error %s", txHash.ToHexString(), overlay.Error())
		}
		if err != nil {
			log.Debugf("HandleInvokeTransaction tx %s error %s", txHash.ToHexString(), err)
		}
	}

	return notify, nil
}

func refreshGlobalParam(gasTable map[string]uint64) {
	for k, v := range gasTable {
		neovm.GAS_TABLE.Store(k, v)
	}
}
