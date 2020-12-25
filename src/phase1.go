package main

import (
)

type Ofst int64
type Amnt int64

const (
    blkStartHeight = 0  //it is fixed now! do not change
    BlockLimit  = 2238270
    BlockHeightofPaper = 1240000

    numIter = 5

    blkHeight = BlockHeightofPaper
)

func Phase1(tb *TracingBlocks) {
    var zero_mixin, traced_txin, total_txin int64 = 0, 0, 0

    var RingCTSpent     map[Ofst]bool           = make(map[Ofst]bool)
    var NonRingCTSpent  map[Amnt]map[Ofst]bool  = make(map[Amnt]map[Ofst]bool)

    var totalti []int = make([]int, blkHeight+1)
    var totaltracedti []int = make([]int, blkHeight+1)

    // --- tracing inputs
    for iterBC:=0; iterBC<numIter; iterBC++ {
        loggerI.Printf("%d'th iteration.\n", iterBC)

        var i int32
        for i=blkStartHeight ; i<blkHeight ; i++ {
            block := tb.GetBlock(i)

            for _, ti := range block.TxInputs {
                if ti.Version == 1 {
                    for j:=0; j<len(ti.Amounts); j++ { //each txin_v
                        var untraced_offsets []Ofst = make([]Ofst, 0, len(ti.Amounts))
                        var amnt Amnt = Amnt(ti.Amounts[j])

                        if iterBC==0 {
                            if total_txin++; len(ti.Goffsetss[j])==1 {
                                zero_mixin++
                            }
                        }

                        for _, offset_r := range ti.Goffsetss[j] {
                            offset := Ofst(offset_r)
                            if _, ok := NonRingCTSpent[amnt][offset]; !ok{ //seen?
                                untraced_offsets = append(untraced_offsets, offset)
                            }
                        }

                        if len(untraced_offsets) == 1 {
                            if _, ok := NonRingCTSpent[amnt]; !ok{
                                NonRingCTSpent[amnt] = make(map[Ofst]bool)
                            }
                            NonRingCTSpent[amnt][untraced_offsets[0]] = true
                            traced_txin++
                        }

                        if iterBC==numIter-1 { //version 1.2
                            totaltracedti[i] += len(ti.Amounts) - len(untraced_offsets)
                            totalti[i] += len(ti.Amounts)
                        }
                    }
                } else if ti.Version == 2 {   // RingCT transaction
                    if iterBC==0 {
                        total_txin += int64(len(ti.Goffsetss))
                    }

                } else {
                    loggerD.Println("other transaction version exist")
                }
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

    /*
    // amounts and offsets
    for amnt,ofstMap := range NonRingCTSpent {
        for offset,_ := range ofstMap {
            r := make([]string, 0, len(ofstMap)+1)
            r = append(r, strconv.FormatInt(int64(amnt),10), strconv.FormatInt(int64(offset),10))
            if err := w_ao.Write(r); err!=nil{
                loggerE.Println(err)
            }
        }
    }

    // traced transaction inputs
    for txHash, indecies := range TracedTxInputs {
        r := make([]string, 0, len(indecies)+1)
        buf := []string(strings.Fields(strings.Trim(fmt.Sprint(indecies),"[]")))
        r = append(r, txHash)
        r = append(r, buf...)
        if err := w_tx.Write(r); err!=nil{
            loggerE.Println(err)
        }
    }
    */
}

