package common

import (
	sgo "github.com/SolmateDev/solana-go"
	sgorpc "github.com/SolmateDev/solana-go/rpc"
	sgows "github.com/SolmateDev/solana-go/rpc/ws"
)

type SlotUpdate struct {
	Result     sgows.SlotResult
	Commitment sgorpc.CommitmentType
}

type TransactionUpdate struct {
	Slot   uint64
	TxId   sgo.Signature
	IsVote bool
	Index  uint64
	Meta   Meta
}

type Meta struct {
	HadErr          bool
	BalancesPre     []uint64
	BalancesPost    []uint64
	ComputeConsumed uint64
}

type AccountUpdate struct {
	TxId      sgo.Signature
	Id        sgo.PublicKey
	Account   sgorpc.Account
	Version   uint64
	Slot      uint64
	Data      []byte
	IsStartup bool
}
