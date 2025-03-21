package models

import "time"

// GenericApiResponse defines the structure of a generic response type
type GenericApiResponse struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
	Code  string      `json:"code"`
}

type NetworkConfig struct {
	NumMetachainNodes  uint64 `json:"klv_num_metachain_nodes"`
	ConsensusGroupSize uint64 `json:"klv_consensus_group_size"`
	ChainID            string `json:"klv_chain_id"`
	SlotDuration       uint64 `json:"klv_slot_duration"`
	SlotsPerEpoch      uint64 `json:"klv_slots_per_epoch"`
	StartTime          uint64 `json:"klv_start_time"`
}

type NetworkConfigResponseData struct {
	NetworkConfig *NetworkConfig `json:"config"`
}

// NetworkConfigResponse defines the structure of responses on NetworkConfigResponse API endpoint
type NetworkConfigResponse struct {
	Data  NetworkConfigResponseData `json:"data"`
	Error string                    `json:"error"`
	Code  string                    `json:"code"`
}

type NodeOverview struct {
	BaseTxSize           int64         `json:"baseTxSize"`
	ChainID              string        `json:"chainID"`
	CurrentSlot          uint64        `json:"currentSlot"`
	EpochNumber          int64         `json:"epochNumber"`
	Nonce                uint64        `json:"nonce"`
	NonceAtEpochStart    uint64        `json:"nonceAtEpochStart"`
	SlotAtEpochStart     uint64        `json:"slotAtEpochStart"`
	SlotCurrentTimestamp time.Duration `json:"slotCurrentTimestamp"`
	SlotDuration         time.Duration `json:"slotDuration"`
	SlotsPerEpoch        uint64        `json:"slotsPerEpoch"`
	StartTime            time.Duration `json:"startTime"`
}

type NodeOverviewResponseData struct {
	NodeOverview *NodeOverview `json:"overview"`
}

type NodeOverviewApiResponse struct {
	Data  NodeOverviewResponseData `json:"data"`
	Error string                   `json:"error"`
	Code  string                   `json:"code"`
}
