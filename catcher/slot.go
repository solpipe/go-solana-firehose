package catcher

import (
	"encoding/binary"
	"errors"
	"io"
	"log"

	sgorpc "github.com/SolmateDev/solana-go/rpc"
	sgows "github.com/SolmateDev/solana-go/rpc/ws"
	"github.com/solpipe/go-solana-firehose/common"
)

// format: [slot, uint64][parent,uint64][status,u8 (0=processed,1=rooted,2=confirmed)]
func (e1 external) parseSlot(c io.Reader) error {
	log.Print("parsing slot")
	doneC := e1.ctx.Done()
	var err error
	result := new(sgows.SlotResult)
	err = binary.Read(c, binary.BigEndian, &result.Slot)
	if err != nil {
		return err
	}
	err = binary.Read(c, binary.BigEndian, &result.Parent)
	if err != nil {
		return err
	}
	var statusN uint8
	err = binary.Read(c, binary.BigEndian, &statusN)
	if err != nil {
		return err
	}

	var confirmationStatus sgorpc.CommitmentType
	switch statusN {
	case 0:
		confirmationStatus = sgorpc.CommitmentProcessed
	case 1:
		confirmationStatus = sgorpc.CommitmentConfirmed
	case 2:
		confirmationStatus = sgorpc.CommitmentFinalized
	default:
		err = errors.New("unknown slot status")
	}
	if err != nil {
		return err
	}

	select {
	case <-doneC:
		err = errors.New("canceled")
	case e1.slotUpdateC <- common.SlotUpdate{
		Result:     *result,
		Commitment: confirmationStatus,
	}:
	}
	if err != nil {
		return err
	}

	return nil
}
