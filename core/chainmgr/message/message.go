package message

import (
	"encoding/json"
	"fmt"
)

const (
	SHARD_PROTOCOL_VERSION = 1
)

const (
	HELLO_MSG = iota
	CONFIG_MSG
	BLOCK_REQ_MSG
	BLOCK_RSP_MSG
	PEERINFO_REQ_MSG
	PEERINFO_RSP_MSG

	DISCONNECTED_MSG
)

type RemoteShardMsg interface {
	Type() int
}

type ShardHelloMsg struct {
	TargetShardID uint64 `json:"target_shard_id"`
	SourceShardID uint64 `json:"source_shard_id"`
}

func (msg *ShardHelloMsg) Type() int {
	return HELLO_MSG
}

type ShardConfigMsg struct {
	Account []byte `json:"account"`
	Config  []byte `json:"config"`

	// peer pk : ip-addr/port, (query ip-addr from p2p)
	// genesis config
}

func (msg *ShardConfigMsg) Type() int {
	return CONFIG_MSG
}

type ShardBlockReqMsg struct {
	ShardID  uint64 `json:"shard_id"`
	BlockNum uint64 `json:"block_num"`
}

func (msg *ShardBlockReqMsg) Type() int {
	return BLOCK_REQ_MSG
}

type ShardBlockRspMsg struct {
	FromShardID uint64            `json:"from_shard_id"`
	Height      uint64            `json:"height"`
	BlockHeader *ShardBlockHeader `json:"block_header"`
	Txs         []*ShardBlockTx   `json:"txs"`
}

func (msg *ShardBlockRspMsg) Type() int {
	return BLOCK_RSP_MSG
}

type ShardGetPeerInfoReqMsg struct {
	PeerPubKey []byte `json:"peer_pub_key"`
}

func (msg *ShardGetPeerInfoReqMsg) Type() int {
	return PEERINFO_REQ_MSG
}

type ShardGetPeerInfoRspMsg struct {
	PeerPubKey  []byte `json:"peer_pub_key"`
	PeerAddress string `json:"peer_address"`
}

func (msg *ShardGetPeerInfoRspMsg) Type() int {
	return PEERINFO_RSP_MSG
}

type ShardDisconnectedMsg struct {
	Address string `json:"address"`
}

func (msg *ShardDisconnectedMsg) Type() int {
	return DISCONNECTED_MSG
}

func DecodeShardMsg(msgtype int32, msgPayload []byte) (RemoteShardMsg, error) {
	switch msgtype {
	case HELLO_MSG:
		msg := &ShardHelloMsg{}
		if err := json.Unmarshal(msgPayload, msg); err != nil {
			return nil, fmt.Errorf("unmarshal remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	case CONFIG_MSG:
		msg := &ShardConfigMsg{}
		if err := json.Unmarshal(msgPayload, msg); err != nil {
			return nil, fmt.Errorf("unmarshal remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	case BLOCK_REQ_MSG:
		msg := &ShardBlockReqMsg{}
		if err := json.Unmarshal(msgPayload, msg); err != nil {
			return nil, fmt.Errorf("unmarshal remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	case BLOCK_RSP_MSG:
		msg := &ShardBlockRspMsg{}
		if err := json.Unmarshal(msgPayload, msg); err != nil {
			return nil, fmt.Errorf("unmarshal remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	case PEERINFO_REQ_MSG:
		msg := &ShardGetPeerInfoReqMsg{}
		if err := json.Unmarshal(msgPayload, msg); err != nil {
			return nil, fmt.Errorf("unmarshal remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	case PEERINFO_RSP_MSG:
		msg := &ShardGetPeerInfoRspMsg{}
		if err := json.Unmarshal(msgPayload, msg); err != nil {
			return nil, fmt.Errorf("unmarshal remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	}
	return nil, fmt.Errorf("unknown remote shard msg type: %d", msgtype)
}
