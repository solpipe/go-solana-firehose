package clock

import (
	"context"
	"errors"

	sgorpc "github.com/SolmateDev/solana-go/rpc"
	"github.com/solpipe/go-solana-firehose/common"
)

type Clock struct {
	ctx  context.Context
	reqC chan<- requestClock
}

type requestClock struct {
	commitment sgorpc.CommitmentType
	errorC     chan<- error
	respC      chan<- common.SlotUpdate
}

func Create(
	ctx context.Context,
	slotC <-chan common.SlotUpdate,
) Clock {
	reqC := make(chan requestClock)
	go loopInternal(
		ctx,
		slotC,
		reqC,
	)

	return Clock{
		ctx:  ctx,
		reqC: reqC,
	}
}

// get the latest slot update
func (e1 Clock) Update(ctx context.Context, commitment sgorpc.CommitmentType) (su common.SlotUpdate, err error) {
	finishC := ctx.Done()
	doneC := e1.ctx.Done()
	errorC := make(chan error, 1)
	respC := make(chan common.SlotUpdate, 1)
	select {
	case <-finishC:
		err = errors.New("canceled")
	case <-doneC:
		err = errors.New("canceled")
	case e1.reqC <- requestClock{
		commitment: commitment,
		errorC:     errorC,
		respC:      respC,
	}:
	}
	if err != nil {
		return
	}
	select {
	case <-finishC:
		err = errors.New("canceled")
	case <-doneC:
		err = errors.New("canceled")
	case err = <-errorC:
	}
	if err != nil {
		return
	}
	su = <-respC
	return
}
