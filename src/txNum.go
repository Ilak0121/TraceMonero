package main

import (
    "fmt"
    "log"
    "os"
    "strings"
    "encoding/csv"
)

var numIter int = 1 //5
var blkHeight int32 = BlockLimit

func main() {
    var i int32

    var Num_zero_mixin int64 = 0
    var Num_total_txin int64 = 0

    var txs_per_block []int64 = make([]int64,0,blkHeight+1)

    f, err := os.OpenFile("./log/txNum.log",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0644)
    if err!=nil{
        log.Fatal(err)
    }

    loggerI := log.New(f, "[INFO] ", log.LstdFlags)
    loggerE := log.New(f, "[ERROR] ", log.LstdFlags)
    //loggerD := log.New(f, "[DEBUG] ", log.LstdFlags)
    log.SetFlags(log.LstdFlags | log.Lshortfile)

    for iterBC:=0; iterBC<numIter; iterBC++ {
        loggerI.Printf("%d'th iteration.\n", iterBC)
        var progress int32 = 0
        for i=0 ; i<blkHeight+1 ; i++ {
            //fmt.Printf("\t[Info] %d'th block.\n", i)
            txHashes := NCBTxsFromBlock(i, loggerE)
            txInfos := GetTxInputInfo(txHashes, loggerE) // []*txinfos : {Version, TxHash, Amounts, Goffsetss}

            for _, txInfo := range txInfos {
                if txInfo.Version == 1 {
                    if len(txInfo.Amounts) != len(txInfo.Goffsetss) {
                        loggerE.Println("len of amounts and goffsetss are different")
                        os.Exit(-1)
                    }

                    // survey for each txin_v
                    for j:=0; j<len(txInfo.Amounts); j++ {
                        if Num_total_txin++; len(txInfo.Goffsetss[j])==1 {
                            Num_zero_mixin++
                        }
                    }
                //} else if txInfo.Version == 2 {   //pass at this time
                } else {
                    loggerE.Println("other transaction version")
                    os.Exit(-1)
                }
            }

            if i==0 {
                txs_per_block = append(txs_per_block, Num_total_txin)
            } else {
                txs_per_block = append(txs_per_block, Num_total_txin - txs_per_block[i-1])
            }

            if progress == (i/20000) { //logging progress
                progress++
                loggerI.Printf("\tprogress of a iter.: %d\n",i)
            }

        }// end 2nd for
    }
    loggerI.Println("*Program Completed*...")
    loggerI.Println("# of total txins :",Num_total_txin)
    loggerI.Println("# of zero mix-ins :",Num_zero_mixin)

    // ---
    f_tx, err := os.OpenFile("./log/txNum.csv",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0644)
    if err!=nil{
        loggerE.Println(err)
        os.Exit(-1)
    }
    defer f_tx.Close()
    w_tx := csv.NewWriter(f_tx)
    defer w_tx.Flush()

    r := make([]string, 0, blkHeight+1)
    r = append(r, string("txs_per_block"))
    r = append(r,[]string(strings.Fields(strings.Trim(fmt.Sprint(txs_per_block),"[]")))...)
    if err := w_tx.Write(r); err!=nil {
        loggerE.Println(err)
    }

    defer f.Close()

}
