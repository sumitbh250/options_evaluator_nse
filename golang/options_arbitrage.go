package main

// var WG sync.WaitGroup
// var ResultMutex sync.Mutex
// var Results []*Result
// var Web *WebSession

// func init() {
// 	Web = NewWebSession(HttpTimeoutSecs, HttpRetryCount, MaxFileDescriptors)
// }
// func main() {
// 	WG.Add(len(OptionsIndiceLotDict) + len(OptionsStockLotDict))
// 	for stock, lotSize := range OptionsIndiceLotDict {
// 		go func(stock string, lotSize float64) {
// 			defer WG.Done()
// 			getStockArbitrageResult(stock, lotSize)
// 		}(stock, lotSize)
// 	}
// 	for stock, lotSize := range OptionsStockLotDict {
// 		go func(stock string, lotSize float64) {
// 			defer WG.Done()
// 			getStockArbitrageResult(stock, lotSize)
// 		}(stock, lotSize)
// 	}

// 	go func() {
// 		for {
// 			fmt.Println(runtime.NumGoroutine())
// 			time.Sleep(30 * time.Second)
// 		}
// 	}()
// 	go func() {
// 		log.Println(http.ListenAndServe("localhost:6060", nil))
// 	}()
// 	WG.Wait()
// }
