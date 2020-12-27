package main

import (
    "sync"
)

const (
    blkStartHeight = 0  //it is fixed now! do not change
    blkHeight = BlockHeightofPaper

    numIter = 5
)

func Phase1(tb *TracingBlocks) {
    var zero_mixin, traced_txin, total_txin, total_tx int64 = 0, 0, 0, 0

    TXSpent := make(map[Pair]bool) //Amnt, Ofst

    totalti := make([]int, blkHeight+1)
    totaltracedti := make([]int, blkHeight+1)

    // --- tracing inputs
    for iterBC:=0; iterBC<numIter; iterBC++ {
        loggerI.Printf("%d'th iteration.\n", iterBC)

        //var apercent int32 = blkHeight/int32(100)
        //var progress int32 = int32(0)

        var wg sync.WaitGroup

        for i:=int32(blkStartHeight) ; i<blkHeight ; i++ {
            /*if iterBC == 0 && progress == i/apercent {
                loggerI.Printf("one iteration progress: %d%%...\n", progress)
                progress++
            }*/

            block := tb.GetBlock(i)

            flag := false
            for _, ti := range block.TxInputs {
                if iterBC == 0 {
                    total_tx++
                }

                if ti.IsCoinbase == true {                      // coinbase has no input
                    continue
                }

                if ti.Version == 1 || ti.Version == 2 {
                    for j:=0; j<len(ti.Amounts); j++ {          // each txin_v
                        untraced_offsets := make([]int64, 0, len(ti.Amounts))
                        amnt := ti.Amounts[j]

                        if iterBC==0 {
                            if total_txin++; len(ti.Goffsetss[j])==1 {
                                zero_mixin++
                            }
                        } else if iterBC==numIter-1 {           // coalessing to be implemented
                            totaltracedti[i] += len(ti.Amounts) - len(untraced_offsets)
                            totalti[i] += len(ti.Amounts)
                        }

                        for _, offset := range ti.Goffsetss[j] {
                            if _, ok := TXSpent[Pair{amnt, offset}]; !ok{ //seen?
                                untraced_offsets = append(untraced_offsets, offset)
                            }
                        }

                        if len(untraced_offsets) == 1 {
                            TXSpent[Pair{amnt, untraced_offsets[0]}] = true
                            traced_txin++
                            if ti.Roffsets[j] != untraced_offsets[0] {
                                ti.Roffsets[j] = untraced_offsets[0]
                                flag = true
                            }
                        }
                    }

                } else {
                    loggerD.Println("other transaction version exist")
                }

            }
            if flag==true {
                loggerD.Printf("%d block update", i)
                wg.Add(1)
                go func(){
                    defer wg.Done()
                    tb.UpdateBlock(i, block)
                }()
            }

            wg.Wait()
        }// end one blockchain
    } // end iterBC

    TXSpent = nil

    if err:=CSVWrite(totalti, TotalInputsFile); err!=nil {
        loggerE.Println(err)
    }
    if err:=CSVWrite(totaltracedti, TotalTracedInputsFile); err!=nil {
        loggerE.Println(err)
    }

    loggerI.Println("** Phase1 Completed **")
    loggerI.Println("# of total tx :",total_tx)
    loggerI.Println("# of total txins :",total_txin)
    loggerI.Println("# of zero mix-ins :",zero_mixin)
    loggerI.Println("# of total traced txins (effective 0 mix-in) :",traced_txin)

    return
}

