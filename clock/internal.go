package clock

import (
	"context"
	"errors"

	sgorpc "github.com/SolmateDev/solana-go/rpc"
	"github.com/solpipe/go-solana-firehose/common"
)

type internal struct {
	ctx context.Context

	confirmed *common.SlotUpdate
	finalized *common.SlotUpdate
	processed *common.SlotUpdate
}

func loopInternal(
	ctx context.Context,
	slotC <-chan common.SlotUpdate,
	reqC <-chan requestClock,
) {
	doneC := ctx.Done()
	in := new(internal)
	in.ctx = ctx
	var su common.SlotUpdate

out:
	for {
		select {
		case <-doneC:
			break out
		case su = <-slotC:
			switch su.Commitment {
			case sgorpc.CommitmentProcessed:
				in.processed = &su
			case sgorpc.CommitmentConfirmed:
				in.confirmed = &su
			case sgorpc.CommitmentFinalized:
				in.finalized = &su
			default:
			}
		case req := <-reqC:
			var suref *common.SlotUpdate
			switch req.commitment {
			case sgorpc.CommitmentProcessed:
				suref = in.processed
			case sgorpc.CommitmentConfirmed:
				suref = in.confirmed
			case sgorpc.CommitmentFinalized:
				suref = in.finalized
			default:
				req.errorC <- errors.New("unknown commitment")
			}
			if suref != nil {
				req.errorC <- nil
				req.respC <- *suref
			} else {
				req.errorC <- errors.New("no commitment")
			}
		}
	}
}
