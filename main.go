package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"test_BitmediaLabs/core/etherscan"
	"test_BitmediaLabs/core/handler"
	"test_BitmediaLabs/core/settings"
	"test_BitmediaLabs/core/transactions"
)

func main() {
	conf, err := settings.Init()
	if err != nil {
		log.Println(err)
		return
	}
	ctx, cancelFunc := context.WithCancel(context.Background())

	stopSignal := interruptListener()

	db, err := connectToMongoDatabase(ctx, conf.MongoDB)
	if err != nil {
		log.Println(err)
		return
	}
	txStorage := transactions.NewMongoStorage(db)

	blockScanner := etherscan.NewScanner(conf.EtherscanAPI, txStorage)
	blockScanner.Start(ctx)

	router := handler.Init(txStorage)
	go func() {
		err = http.ListenAndServe(conf.HTTP.URL(), router)
		if err != nil {
			log.Println("failed to start http api")
			return
		}
	}()

	for {
		select {
		case <-stopSignal:
			cancelFunc()
			return
		}
	}

}


func interruptListener() <-chan struct{} {
	var interruptSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}

	done := make(chan struct{})
	go func() {
		interruptChannel := make(chan os.Signal, 1)
		signal.Notify(interruptChannel, interruptSignals...)

		select {
		case sig := <-interruptChannel:
			log.Println("Received signal " + sig.String() + ". Shutting down...")
		}
		close(done)

		for {
			select {
			case sig := <-interruptChannel:
				log.Println("Received signal " + sig.String() + ". Already shutting down...")
			}
		}
	}()

	return done
}