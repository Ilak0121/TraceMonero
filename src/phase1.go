package main

import (
    "reflect"
)

type Ofst int64
type Amnt int64

const (
    blkStartHeight = 0  //it is fixed now! do not change

    numIter = 5
    blkHeight = BlockHeightofPaper
)

func Phase1(tb *TracingBlocks) {
    var zero_mixin, traced_txin, total_txin int64 = 0, 0, 0

    var TXSpent  map[Amnt]map[Ofst]bool = make(map[Amnt]map[Ofst]bool)

    var totalti []int = make([]int, blkHeight+1)
    var totaltracedti []int = make([]int, blkHeight+1)

    // --- tracing inputs
    for iterBC:=0; iterBC<numIter; iterBC++ {
        loggerI.Printf("%d'th iteration.\n", iterBC)

        var apercent int32 = blkHeight/int32(100)
        var progress int32 = int32(0)
        for i:=int32(blkStartHeight) ; i<blkHeight ; i++ {
            if iterBC == 0 && progress == i/apercent {
                loggerI.Printf("one iteration progress: %d%%...\n", progress)
                progress++
            }

            block := tb.GetBlock(i)
            updateFlag:= false

            for _, ti := range block.TxInputs {
                if ti.IsCoinbase == true {                      // coinbase has no input
                    continue
                }

                var roffsets []int64 = make([]int64, len(ti.Goffsetss))

                if ti.Version == 1 || ti.Version == 2 {
                    for j:=0; j<len(ti.Amounts); j++ {          // each txin_v
                        var untraced_offsets []Ofst = make([]Ofst, 0, len(ti.Amounts))
                        var amnt Amnt = Amnt(ti.Amounts[j])

                        if iterBC==0 {
                            if total_txin++; len(ti.Goffsetss[j])==1 {
                                zero_mixin++
                            }
                        } else if iterBC==numIter-1 {           // coalessing to be implemented
                            totaltracedti[i] += len(ti.Amounts) - len(untraced_offsets)
                            totalti[i] += len(ti.Amounts)
                        }

                        for _, offset_r := range ti.Goffsetss[j] {
                            offset := Ofst(offset_r)
                            if _, ok := TXSpent[amnt][offset]; !ok{ //seen?
                                untraced_offsets = append(untraced_offsets, offset)
                            }
                        }

                        if len(untraced_offsets) == 1 {
                            if _, ok := TXSpent[amnt]; !ok{
                                TXSpent[amnt] = make(map[Ofst]bool)
                            }
                            TXSpent[amnt][untraced_offsets[0]] = true
                            traced_txin++
                            roffsets[j] = int64(untraced_offsets[0])
                        }
                    }

                } else {
                    loggerD.Println("other transaction version exist")
                }

                if reflect.DeepEqual(roffsets,ti.Roffsets) == false { //block roffset update
                    ti.Roffsets = roffsets
                    updateFlag = true
                }
            }
            if updateFlag==true {
                go tb.UpdateBlock(i, block)
            }

        }// end one blockchain
    } // end iterBC

    loggerI.Println("** Phase1 Completed **")
    loggerI.Println("# of total txins :",total_txin)
    loggerI.Println("# of zero mix-ins :",zero_mixin)
    loggerI.Println("# of total traced txins (effective 0 mix-in) :",traced_txin)

    // version 1.2
    if err:=CSVWrite(totalti, TotalInputsFile); err!=nil {
        loggerE.Println(err)
    }
    if err:=CSVWrite(totaltracedti, TotalTracedInputsFile); err!=nil {
        loggerE.Println(err)
    }

    return
}

