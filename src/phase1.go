package main

import (
    "fmt"
    "log"
    "os"
    "strings"
    "strconv"
    "encoding/csv"
)

type Ofst int64
type Amnt int64

var numIter int = 3 //5
var blkHeight int32 = BlockLimit

func main() {
    var i int32

    var Num_zero_mixin int64 = 0
    var Num_traced_txin int64 = 0
    var Num_total_txin int64 = 0

    var txs_per_block []int64 = make([]int64,0,blkHeight+1)
    var ttxs_per_block []int64 = make([]int64,0,blkHeight+1)

    //var RingCTSpent     map[Ofst]bool
    var NonRingCTSpent  map[Amnt]map[Ofst]bool
    var TracedTxInputs map[string][]int32

    //RingCTSpent = make(map[Ofst]bool)
    NonRingCTSpent = make(map[Amnt]map[Ofst]bool)
    TracedTxInputs = make(map[string][]int32)

    f, err := os.OpenFile("phase1.log",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0644)
    if err!=nil{
        log.Fatal(err)
    }
    defer f.Close()

    loggerI := log.New(f, "[INFO]", log.LstdFlags)
    loggerE := log.New(f, "[ERROR]", log.LstdFlags)
    //loggerD := log.New(f, "[DEBUG]", log.LstdFlags)
    log.SetFlags(log.LstdFlags | log.Lshortfile)

    for iterBC:=0; iterBC<numIter; iterBC++ {
        loggerI.Printf("%d'th iteration.\n", iterBC)
        var progress int32 = 0
        for i=0 ; i<blkHeight+1 ; i++ {
            //fmt.Printf("\t[Info] %d'th block.\n", i)
            txHashes := NCBTxsFromBlock(i)
            txInfos := GetTxInputInfo(txHashes) // []*txinfos : {Version, TxHash, Amounts, Goffsetss}

            for _, txInfo := range txInfos {
                if txInfo.Version == 1 {
                    if len(txInfo.Amounts) != len(txInfo.Goffsetss) {
                        loggerE.Println("len of amounts and goffsetss are different")
                        os.Exit(-1)
                    }

                    // survey for each txin_v
                    for j:=0; j<len(txInfo.Amounts); j++ {
                        var untraced_offsets []Ofst
                        var txAmount Amnt = Amnt(txInfo.Amounts[j])

                        if iterBC==0 {
                            if Num_total_txin++; len(txInfo.Goffsetss[j])==1 {
                                Num_zero_mixin++
                            }
                        }

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
                            Num_traced_txin++
                            //fmt.Println("[INFO] tx, data, idx :",TracedTxInputs[string(txInfo.TxHash)], NonRingCTSpent[txAmount], idx)
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
                ttxs_per_block = append(ttxs_per_block, Num_traced_txin)
            } else {
                txs_per_block = append(txs_per_block, Num_total_txin - txs_per_block[i-1])
                ttxs_per_block = append(ttxs_per_block, Num_traced_txin - ttxs_per_block[i-1])
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
    loggerI.Println("# of total traced txins (effective 0 mix-in) :",Num_traced_txin)

    // ---
    f_ao, err := os.OpenFile("amnt_ofst.csv",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0644)
    if err!=nil{
        loggerE.Println(err)
        os.Exit(-1)
    }
    defer f_ao.Close()
    w_ao := csv.NewWriter(f_ao)
    defer w_ao.Flush()

    for amnt,ofstMap := range NonRingCTSpent {
        for offset,_ := range ofstMap {
            r := make([]string, 0, len(ofstMap)+1)
            r = append(r, strconv.FormatInt(int64(amnt),10), strconv.FormatInt(int64(offset),10))
            if err := w_ao.Write(r); err!=nil{
                loggerE.Println(err)
            }
        }
    }
    // ---
    f_tx, err := os.OpenFile("spent_txs.csv",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0644)
    if err!=nil{
        loggerE.Println(err)
        os.Exit(-1)
    }
    defer f_tx.Close()
    w_tx := csv.NewWriter(f_tx)
    defer w_tx.Flush()

    for txHash, indecies := range TracedTxInputs {
        r := make([]string, 0, len(indecies)+1)
        buf := []string(strings.Fields(strings.Trim(fmt.Sprint(indecies),"[]")))
        r = append(r, txHash)
        r = append(r, buf...)
        if err := w_tx.Write(r); err!=nil{
            loggerE.Println(err)
        }
    }

    r := make([]string, 0, blkHeight+1)
    r = append(r, string("txs_per_block"))
    r = append(r,[]string(strings.Fields(strings.Trim(fmt.Sprint(txs_per_block),"[]")))...)
    if err := w_tx.Write(r); err!=nil {
        loggerE.Println(err)
    }

    r = r[:0]
    r = append(r, string("ttxs_per_block"))
    r = append(r,[]string(strings.Fields(strings.Trim(fmt.Sprint(ttxs_per_block),"[]")))...)
    if err := w_tx.Write(r); err!=nil {
        loggerE.Println(err)
    }

}
