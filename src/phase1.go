package main

import (
    "fmt"
    "log"
    //"os"
)

type Ofst int64
type Amnt int64

var numIter int = 2 //5

func main() {

    log.SetFlags(log.LstdFlags | log.Lshortfile)

    var i int32

    //var RingCTSpent     map[Ofst]bool
    var NonRingCTSpent  map[Amnt]map[Ofst]bool
    var TracedTxInputs map[string][]int32

    NonRingCTSpent = make(map[Amnt]map[Ofst]bool)
    TracedTxInputs = make(map[string][]int32)

    for iterBlock:=0; iterBlock<numIter; iterBlock++ {
        fmt.Printf("[INFO] %d'th iteration.\n", iterBlock)
        for i=0 ; i<2000 ; i++ {
            //fmt.Printf("\t[Info] %d'th block.\n", i)
            txHashes := NCBTxsFromBlock(i)
            txInfos := GetTxInputInfo(txHashes) // []*txinfos : {Version, TxHash, Amounts, Goffsetss}

            for _, txInfo := range txInfos {
                if txInfo.Version == 1 {
                    if len(txInfo.Amounts) != len(txInfo.Goffsetss) {
                        log.Fatal("[ERROR] len of amounts and goffsetss are different")
                    }

                    // survey for each txin_v
                    for j:=0; j<len(txInfo.Amounts); j++ {
                        var untraced_offsets []Ofst
                        var txAmount Amnt = Amnt(txInfo.Amounts[j])

                        for _, offset_r := range (txInfo.Goffsetss[j]) {
                            offset := Ofst(offset_r)
                            if _, ok := NonRingCTSpent[txAmount][offset]; !ok{ //seen?
                                untraced_offsets = append(untraced_offsets, offset)
                            }
                        }
                        if len(untraced_offsets) == 1 {
                            TracedTxInputs[string(txInfo.TxHash)] = append(TracedTxInputs[string(txInfo.TxHash)], int32(j))
                            if _, ok := NonRingCTSpent[txAmount]; !ok{
                                NonRingCTSpent[txAmount] = make(map[Ofst]bool)
                            }
                            NonRingCTSpent[txAmount][untraced_offsets[0]] = true
                            //fmt.Println("[INFO] tx, data, idx :",TracedTxInputs[string(txInfo.TxHash)], NonRingCTSpent[txAmount], idx)
                        }
                    }
                } else if txInfo.Version == 2 {
                    //pass at this time
                } else {
                    log.Fatal("[ERROR] other transaction version")
                }
            }
        }
    }
    fmt.Println("[INFO] Done...")
    fmt.Println("[INFO] Traced Data")
    for k, v := range NonRingCTSpent {
        fmt.Printf("\t")
        fmt.Println("Traced amount & offset : ",k,v)
    }
    fmt.Println("[INFO] Traced Txs")
    for k, v := range TracedTxInputs{
        fmt.Printf("\t")
        fmt.Println("Traced Txhash & input_index : ",k,v)
    }
}
