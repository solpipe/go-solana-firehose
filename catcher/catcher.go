package catcher

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	sgo "github.com/SolmateDev/solana-go"
	"github.com/solpipe/go-solana-firehose/common"
)

type external struct {
	ctx                     context.Context
	cancel                  context.CancelFunc
	account_socket_path     string
	transaction_socket_path string
	slot_socket_path        string
	errorC                  chan<- error
	slotUpdateC             chan<- common.SlotUpdate
	accountUpdateC          chan<- common.AccountUpdate
	transactionUpdateC      chan<- common.TransactionUpdate
}

type ReceiverGroup struct {
	ErrorC       <-chan error
	SlotC        <-chan common.SlotUpdate
	AccountC     <-chan common.AccountUpdate
	TransactionC <-chan common.TransactionUpdate
}

func RunDetached(
	ctx context.Context,
	directory string,
	bufferSize uint,
	workers uint,
) (rg ReceiverGroup) {

	ctxC, cancel := context.WithCancel(ctx)
	errorC := make(chan error, 1)
	summaryErrorC := make(chan error, 1)
	slotUpdateC := make(chan common.SlotUpdate)
	accountUpdateC := make(chan common.AccountUpdate)
	transactionUpdateC := make(chan common.TransactionUpdate)
	rg = ReceiverGroup{
		ErrorC:       summaryErrorC,
		SlotC:        slotUpdateC,
		AccountC:     accountUpdateC,
		TransactionC: transactionUpdateC,
	}
	e1 := external{
		ctx:                     ctxC,
		cancel:                  cancel,
		errorC:                  errorC,
		slotUpdateC:             slotUpdateC,
		accountUpdateC:          accountUpdateC,
		transactionUpdateC:      transactionUpdateC,
		account_socket_path:     fmt.Sprintf("%s/account.socket", directory),
		slot_socket_path:        fmt.Sprintf("%s/slot.socket", directory),
		transaction_socket_path: fmt.Sprintf("%s/transaction.socket", directory),
	}

	go e1.relayError(errorC, summaryErrorC)

	slotC := make(chan net.Conn, bufferSize)
	go e1.listen(e1.slot_socket_path, slotC)
	for i := uint(0); i < workers; i++ {
		go e1.run(slotC, e1.parseSlot)
	}

	txC := make(chan net.Conn, bufferSize)
	go e1.listen(e1.transaction_socket_path, txC)
	for i := uint(0); i < workers; i++ {
		go e1.run(txC, e1.parseTransaction)
	}

	accountC := make(chan net.Conn, bufferSize)
	go e1.listen(e1.account_socket_path, accountC)
	for i := uint(0); i < workers; i++ {
		go e1.run(accountC, e1.parseAccount)
	}

	return
}

func (e1 external) relayError(inC <-chan error, outC chan<- error) {

	doneC := e1.ctx.Done()
	var err error
	select {
	case <-doneC:
	case err = <-inC:
	}
	e1.cancel()
	<-time.After(2 * time.Second)
	outC <- err
}

func (e1 external) run(clientC <-chan net.Conn, cb func(io.Reader) error) {
	doneC := e1.ctx.Done()
	var c net.Conn
	var err error
out:
	for {
		select {
		case <-doneC:
			break out
		case c = <-clientC:
			err = cb(c)
			if err != nil {
				log.Printf("finished callback with err: %s", err.Error())
			} else {
				log.Print("callback ok")
				c.Write([]byte{1})
			}
			c.Close()
		}
	}
	if err != nil {
		e1.errorC <- err
	}
}

func (e1 external) listen(path string, clientC chan<- net.Conn) {
	doneC := e1.ctx.Done()
	os.Remove(path)
	l, err := net.Listen("unix", path)
	if err != nil {
		e1.errorC <- err
		return
	}
	go e1.close(l, path)
out:
	for {
		c, err := l.Accept()
		if err != nil {
			select {
			case <-doneC:
			case e1.errorC <- err:
			}
			break out
		}
		select {
		case <-doneC:
		case clientC <- c:
			log.Print("sending client")
		}
	}
}

func (e1 external) close(l net.Listener, path string) {
	<-e1.ctx.Done()
	l.Close()
}

const (
	SIZE_BOOL int = 1
	SIZE_U8   int = 1
	SIZE_U32  int = 4
	SIZE_U64  int = 8
)

func readBool(c io.Reader) (bool, error) {
	var ans bool
	var n uint8
	err := binary.Read(c, binary.BigEndian, &n)
	if err != nil {
		return false, err
	}
	switch n {
	case 0:
		ans = false
	case 1:
		ans = true
	default:
		return false, errors.New("unknown startup value")
	}

	return ans, nil
}

func readPubkey(c io.Reader) (sgo.PublicKey, error) {
	x, err := readBuf(c, sgo.PublicKeyLength)
	if err != nil {
		return sgo.PublicKey{}, err
	}
	return sgo.PublicKeyFromBytes(x), nil
}

func readBuf(c io.Reader, length int) ([]byte, error) {
	buf := make([]byte, length)
	var n int
	var t int
	var err error

	t = 0
	for t < length {
		n, err = c.Read(buf[t:])
		if err != nil {
			return nil, err
		}
		t += n
	}
	return buf, nil
}

func readSignature(c io.Reader) (sgo.Signature, error) {
	buf, err := readBuf(c, sgo.SignatureLength)
	if err != nil {
		return sgo.Signature{}, err
	}
	return sgo.SignatureFromBytes(buf), nil
}
