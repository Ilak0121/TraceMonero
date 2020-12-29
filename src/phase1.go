package main

import (
    "sync"
    "strconv"
    "github.com/boltdb/bolt"
)

const (
    blkStartHeight = 0  //it is fixed now! do not change
    blkHeight = BlockHeightofPaper

    numIter = 5
)

func phase1(tb *TracingBlocks) {
    var zero_mixin, traced_txin, total_txin, total_tx int64 = 0, 0, 0, 0

    TXSpent := make(map[Pair]bool, 16065185) //Amnt, Ofst

    totalti := make([]int, (blkHeight)/int32(720)+10)
    totaltracedti := make([]int, (blkHeight)/int32(720)+10)

    var count [11][]int64 //0~10
    for i:=0; i<len(count); i++ {
        count[i] = make([]int64, i+1)
    }

    // --- tracing inputs
    for iterBC:=0; iterBC<numIter; iterBC++ {
        loggerI.Printf("%d'th iteration.\n", iterBC)

        //var apercent int32 = blkHeight/int32(100)
        //var progress int32 = int32(0)

        var wg sync.WaitGroup
        var counting int32 = 0
        for i:=int32(blkStartHeight) ; i<blkHeight ; i++ {
            /*if progress == i/apercent {
                if progress++; progress%10==0{
                    loggerI.Printf("one iteration progress: %d%%...\n", progress)
                }
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

                /*if debug == true {
                    var roffset []int64
                    for reset:=0; reset<len(ti.Amounts); reset++{
                        roffset = append(roffset, -1)
                    }
                    ti.Roffsets=roffset
                    flag = true
                    loggerD.Printf("test: %v\n", ti.Roffsets)
                } else {*/

                if ti.Version == 1 || ti.Version == 2 {
                    for j:=0; j<len(ti.Amounts); j++ {          // each txin_v
                        var untraced_offsets []int64
                        amnt := ti.Amounts[j]

                        for _, offset := range ti.Goffsetss[j] {
                            if v, _ := TXSpent[Pair{amnt, offset}]; v!=true { //seen?
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
                        if iterBC==0 {
                            if total_txin++; len(ti.Goffsetss[j])==1 {
                                zero_mixin++
                                //loggerD.Printf("test) traced_txin : %d\n", traced_txin)
                                //loggerD.Printf("test) length      : %d\n", len(untraced_offsets))
                            }
                        } else if iterBC==numIter-1 {           // coalessing to be implemented
                            totaltracedti[i/int32(720)]  += len(ti.Goffsetss[j]) - len(untraced_offsets)
                            totalti[i/int32(720)]        += len(ti.Goffsetss[j])
                        }
                    }
                } else {
                    loggerD.Println("other transaction version exist")
                }

            }
            if flag==true {
                wg.Add(1)
                counting++
                go func(idx int32, bt *BlockTxs){
                    defer wg.Done()
                    err := tb.db.Batch(func(tx *bolt.Tx) error {
                        b := tx.Bucket([]byte(traceBucket))
                        index := strconv.FormatInt(int64(idx),10)
                        err := b.Put([]byte(index), bt.Serialization())
                        if err!=nil {
                            return err
                        }
                        return nil
                    })
                    if err!=nil {
                        loggerE.Println(err)
                    }
                }(i, block)
            }

        }// end one blockchain
        //loggerD.Printf("traced_txin : %d\n", traced_txin)
        //loggerD.Printf("zero_mixin  : %d\n", zero_mixin)
        //loggerD.Printf("# of waiting: %d\n", counting)
        wg.Wait()
    } // end iterBC

    //loggerD.Printf("# of TXSpent: %d\n", len(TXSpent))
    TXSpent = nil

    if err:=CSVWrite(totalti, TotalInputsFile); err!=nil {
        loggerE.Println(err)
    }
    if err:=CSVWrite(totaltracedti, TotalTracedInputsFile); err!=nil {
        loggerE.Println(err)
    }

    loggerI.Println("** Phase1 Completed **")
    loggerI.Println("# of total tx                                :", total_tx)
    loggerI.Println("# of total txins                             :", total_txin)
    loggerI.Println("# of zero mix-ins                            :", zero_mixin)
    loggerI.Println("# of total traced txins (effective 0 mix-in) :", traced_txin)

    return
}

