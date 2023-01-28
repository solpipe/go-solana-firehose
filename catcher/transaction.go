package catcher

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/solpipe/go-solana-firehose/common"
)

// format: [slot,u64][txid, sig][isVote,bool][index,usize][transaction,sanitized]
// [meta,meta]
func (e1 external) parseTransaction(c io.Reader) error {
	doneC := e1.ctx.Done()
	var err error

	var slot uint64
	err = binary.Read(c, binary.BigEndian, &slot)
	if err != nil {
		return err
	}

	txid, err := readSignature(c)
	if err != nil {
		return err
		//return nil
	}
	isVote, err := readBool(c)
	if err != nil {
		return err
	}

	var index uint64
	err = binary.Read(c, binary.BigEndian, &index)
	if err != nil {
		return err
	}

	meta, err := readMeta(c)
	if err != nil {
		return err
	}

	select {
	case <-doneC:
		return errors.New("canceled")
	case e1.transactionUpdateC <- common.TransactionUpdate{
		Slot:   slot,
		TxId:   txid,
		IsVote: isVote,
		Index:  index,
		Meta:   meta,
	}:
	}

	return nil
}

func readMeta(c io.Reader) (m common.Meta, err error) {
	isErr, err := readBool(c)
	if err != nil {
		return
	}
	var preBalList []uint64
	{
		var n uint64
		err = binary.Read(c, binary.BigEndian, &n)
		if err != nil {
			return
		}
		list := make([]uint64, n)
		for i := uint64(0); i < n; i++ {
			err = binary.Read(c, binary.BigEndian, &list[i])
			if err != nil {
				return
			}
		}
		preBalList = list
	}
	var postBalList []uint64
	{
		var n uint64
		err = binary.Read(c, binary.BigEndian, &n)
		if err != nil {
			return
		}
		list := make([]uint64, n)
		for i := uint64(0); i < n; i++ {
			err = binary.Read(c, binary.BigEndian, &list[i])
			if err != nil {
				return
			}
		}
		postBalList = list
	}

	var computeConsumed uint64
	err = binary.Read(c, binary.BigEndian, &computeConsumed)
	if err != nil {
		return
	}

	m = common.Meta{
		HadErr:          isErr,
		BalancesPre:     preBalList,
		BalancesPost:    postBalList,
		ComputeConsumed: computeConsumed,
	}
	return
}
