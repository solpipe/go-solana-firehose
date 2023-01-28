package catcher

import (
	"encoding/binary"
	"errors"
	"io"

	sgorpc "github.com/SolmateDev/solana-go/rpc"
	"github.com/solpipe/go-solana-firehose/common"
)

func (e1 external) parseAccount(c io.Reader) error {
	doneC := e1.ctx.Done()
	var err error

	var slot uint64
	err = binary.Read(c, binary.BigEndian, &slot)
	if err != nil {
		return err
	}

	startup, err := readBool(c)
	if err != nil {
		return err
	}

	txnId, err := readSignature(c)
	if err != nil {
		return err
	}

	accountId, err := readPubkey(c)
	if err != nil {
		return err
	}

	var lamports uint64
	err = binary.Read(c, binary.BigEndian, &lamports)
	if err != nil {
		return err
	}

	ownerId, err := readPubkey(c)
	if err != nil {
		return err
	}

	isExec, err := readBool(c)
	if err != nil {
		return err
	}

	var rent uint64
	err = binary.Read(c, binary.BigEndian, &rent)
	if err != nil {
		return err
	}

	var writeVersion uint64
	err = binary.Read(c, binary.BigEndian, &writeVersion)
	if err != nil {
		return err
	}

	var dataSize uint64
	err = binary.Read(c, binary.BigEndian, &dataSize)
	if err != nil {
		return err
	}

	data, err := readBuf(c, int(dataSize))
	if err != nil {
		return err
	}

	select {
	case <-doneC:
		return errors.New("canceled")
	case e1.accountUpdateC <- common.AccountUpdate{
		TxId:      txnId,
		Slot:      slot,
		Id:        accountId,
		Data:      data,
		IsStartup: startup,
		Version:   writeVersion,
		Account: sgorpc.Account{
			Lamports:   lamports,
			Owner:      ownerId,
			Executable: isExec,
			RentEpoch:  rent,
		},
	}:
	}

	return nil
}
