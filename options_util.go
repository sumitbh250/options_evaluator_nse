
package main

func combinationsWithReplacement(iterable []TradeIfc, r int32) chan []TradeIfc {
	pool := iterable
	n := len(pool)
	if pool == nil || n <= 0 {
		return nil
	}
	indices := make([]int, r)
	chnl := make(chan []TradeIfc)
	go func() {
		defer close(chnl)
		newSlice := make([]TradeIfc, r)
		for i, j := range indices {
			newSlice[i] = pool[j]
		}
		chnl <- newSlice
		for {
			retSlice := make([]TradeIfc, r)
			i := (r-1)
			for ; i >= 0; i-- {
				if indices[i] != n-1 {
					break
				}
			}
			if i < 0 {
				return
			}
			indices[i]++
			for j := i; j < r; j++ {
				indices[j] = indices[i]
			}
			// if indices[0] != 0 || indices[1] != 1 {
			// 	continue
			// }
			// if indices[0] == 1 || indices[1] == 0 {
			// 	continue
			// }
			for j := int32(0); j < r; j++ {
				retSlice[j] = pool[indices[j]]
			}
			chnl <- retSlice
		}
	}()
	return chnl
}
