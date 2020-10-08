package main

import pflow "github.com/UCLabNU/proto_pflow"

// PFlowBlock : type definition of PFlowBlock
type PFlowBlock struct {
	BaseDate int64
	PrevLen  uint32
	PFlows   []*pflow.PFlow
}
