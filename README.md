# Solana Firehose Go Catcher

This Go module listens for account, slot, and transaction updates from a locally running Solana validator.  The IPC takes place over unix domain sockets.  This module relies on the [Solana Firehose Geyser Plugin](https://github.com/solpipe/solana-firehose).

# Usage

See [the Catcher test](catcher/catcher_test.go) for an example of how to use the Firehose catcher.

```go
func DoStuff(ctx context.Context) error{
	ctxC, cancel := context.WithCancel(ctx)
	defer cancel()
	directory := "/tmp/catcher"
	_, err := os.Stat(directory)
	if os.IsNotExist(err) {
		err = os.MkdirAll(directory, 0755)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	rg := catcher.RunDetached(ctx, directory, 10, 3)
	go loopAccount(ctx, rg)
	doneC := ctx.Done()
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
	if err != nil {
		return err
	}
    return nil
}
```