package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

var WG sync.WaitGroup
var ResultMutex sync.Mutex
var Results []*Result
var Web *WebSession
var Bot *tgbotapi.BotAPI

func init() {
	Web = NewWebSession(HttpTimeoutSecs, HttpRetryCount, MaxFileDescriptors)
	var err error
	Bot, err = tgbotapi.NewBotAPI(TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}
	Bot.Debug = true
}

func main() {
	WG.Add(len(OptionsIndiceLotDict) + len(OptionsStockLotDict))
	for stock, lotSize := range OptionsIndiceLotDict {
		go func(stock string, lotSize float64) {
			defer WG.Done()
			getStockResult(stock, lotSize)
		}(stock, lotSize)
	}
	for stock, lotSize := range OptionsStockLotDict {
		go func(stock string, lotSize float64) {
			defer WG.Done()
			getStockResult(stock, lotSize)
		}(stock, lotSize)
	}

	go func() {
		for {
			fmt.Println(runtime.NumGoroutine())
			time.Sleep(30 * time.Second)
		}
	}()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	WG.Wait()
	if ResultAnalysisType == AnalysisType_TERMINAL {
		AnalyzeResults()
	} else if ResultAnalysisType == AnalysisType_TELEGRAM {
		SendResultAnalysisViaTelegram()
	}
}

