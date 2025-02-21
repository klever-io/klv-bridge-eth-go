package models

type NetworkConfig struct {
	NumMetachainNodes  uint64 `json:"klv_num_metachain_nodes"`
	ConsensusGroupSize uint64 `json:"klv_consensus_group_size"`
	ChainID            string `json:"klv_chain_id"`
	SlotDuration       uint64 `json:"klv_slot_duration"`
	SlotsPerEpoch      uint64 `json:"klv_slots_per_epoch"`
	StartTime          uint64 `json:"klv_start_time"`
}

// NetworkConfigResponse defines the structure of responses on NetworkConfigResponse API endpoint
type NetworkConfigResponse struct {
	Data  *NetworkConfig `json:"config"`
	Error string         `json:"error"`
	Code  string         `json:"code"`
}
