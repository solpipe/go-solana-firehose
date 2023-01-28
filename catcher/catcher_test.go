package catcher_test

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/solpipe/go-solana-firehose/catcher"
)

func TestBasic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
	})
	directory := "/tmp/catcher"
	_, err := os.Stat(directory)
	if os.IsNotExist(err) {
		err = os.MkdirAll(directory, 0755)
		if err != nil {
			t.Fatal(err)
		}
	} else if err != nil {
		t.Fatal(err)
	}

	rg := catcher.RunDetached(ctx, directory, 10, 3)
	go loopAccount(ctx, rg)
	t.Cleanup(func() {
		<-time.After(30 * time.Second)
	})
	doneC := ctx.Done()
	select {
	case <-doneC:
		err = errors.New("canceled")
	case err = <-rg.ErrorC:
	case <-time.After(15 * time.Minute):
	}
	if err != nil {
		t.Fatal(err)
	}

}

func loopAccount(
	ctx context.Context,
	rg catcher.ReceiverGroup,
) {
	doneC := ctx.Done()
	i := 0
out:
	for {
		select {
		case <-doneC:
			break out
		case su := <-rg.SlotC:
			log.Printf("slot=%d", su.Result.Slot)
		case x := <-rg.AccountC:
			if i%20 == 0 {
				log.Printf("account=%+v", x)
			}
		case x := <-rg.TransactionC:
			log.Printf("transaction=%+v", x)
		}
		i++
	}
}
