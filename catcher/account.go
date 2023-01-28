package catcher

import (
	"encoding/binary"
	"errors"
	"io"
	"log"

	sgo "github.com/SolmateDev/solana-go"
	sgorpc "github.com/SolmateDev/solana-go/rpc"
)

type AccountUpdate struct {
	TxId      sgo.Signature
	Id        sgo.PublicKey
	Account   sgorpc.Account
	Version   uint64
	Slot      uint64
	Data      []byte
	IsStartup bool
}

// format: [startup, u8/bool][id,pubkey][lamports,u64][owner,pubkey][exec,u8/bool][rent,u64][version,u64][data,?]
func (e1 external) parseAccount(c io.Reader) error {
	log.Print("parsing account")
	doneC := e1.ctx.Done()
	var err error

	var slot uint64
	err = binary.Read(c, binary.BigEndian, &slot)
	if err != nil {
		return err
	}
	log.Printf("...slot=%d", slot)

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
	log.Printf("...account=%s", accountId.String())

	var lamports uint64
	err = binary.Read(c, binary.BigEndian, &lamports)
	if err != nil {
		return err
	}
	log.Printf("...lamports=%d", lamports)
	ownerId, err := readPubkey(c)
	if err != nil {
		return err
	}
	log.Printf("...owner=%s", ownerId.String())

	isExec, err := readBool(c)
	if err != nil {
		return err
	}
	log.Printf("...exec=%v", isExec)

	var rent uint64
	err = binary.Read(c, binary.BigEndian, &rent)
	if err != nil {
		return err
	}
	log.Printf("...rent=%d", rent)

	var writeVersion uint64
	err = binary.Read(c, binary.BigEndian, &writeVersion)
	if err != nil {
		return err
	}
	log.Printf("...writeVersion=%d", writeVersion)

	var dataSize uint64
	err = binary.Read(c, binary.BigEndian, &dataSize)
	if err != nil {
		return err
	}
	log.Printf("...size=%d", dataSize)
	data, err := readBuf(c, int(dataSize))
	if err != nil {
		return err
	}

	select {
	case <-doneC:
		return errors.New("canceled")
	case e1.accountUpdateC <- AccountUpdate{
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
