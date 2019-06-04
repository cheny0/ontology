package TestCommon

import (
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/chainmgr"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/tests"
)

func CreateChain(t *testing.T, name string, shardID common.ShardID, genesisParentHeight uint32) {
	if lgr := ledger.GetShardLedger(shardID); lgr != nil {
		return
	}

	cfg := GetConfig(t, shardID)
	if cfg == nil {
		t.Fatalf("nil config for shard %d", shardID)
	}

	dataDir := TestConsts.TestRootDir + "Chain/" + name + "/"

	var lgr *ledger.Ledger
	var err error
	if shardID.IsRootShard() {
		lgr, err = ledger.NewLedger(dataDir, 0)
	} else {
		rootLgr := ledger.GetShardLedger(shardID.ParentID())
		lgr, err = ledger.NewShardLedger(shardID, dataDir, rootLgr)
	}

	if err != nil {
		t.Fatalf("failed to init ledger %d: %s", shardID, err)
	}

	bookkeeper := GetAccount(chainmgr.GetShardName(shardID) + "_peerOwner0")
	if bookkeeper == nil {
		t.Fatalf("failed to get shard %d, user peer owner0", shardID)
	}
	shardCfg := &config.ShardConfig{
		ShardID:             shardID,
		GenesisParentHeight: genesisParentHeight,
	}
	bookkeepers := []keypair.PublicKey{bookkeeper.PublicKey}
	blk, err := genesis.BuildGenesisBlock(bookkeepers, cfg.Genesis, shardCfg)
	if err != nil {
		t.Fatalf("build genesis block %d, err: %s", shardID, err)
	}

	if err := lgr.Init(bookkeepers, blk); err != nil {
		t.Fatalf("init ledger %d failed: %s", shardID, err)
	}
}

func GetHeight(t *testing.T, shardID common.ShardID) uint32 {
	if lgr := ledger.GetShardLedger(shardID); lgr != nil {
		return lgr.GetCurrentBlockHeight()
	}
	t.Fatalf("get height with invalid shard %d", shardID)
	return 0
}

func GetLastBlock(t *testing.T, shardID common.ShardID) *types.Block {
	if lgr := ledger.GetShardLedger(shardID); lgr != nil {
		blk, err := lgr.GetBlockByHeight(lgr.GetCurrentBlockHeight())
		if err != nil {
			t.Fatalf("get last block of shard %d: %s", shardID, err)
		}
		return blk
	}
	t.Fatalf("get last block with invalid shard %d", shardID)
	return nil
}
