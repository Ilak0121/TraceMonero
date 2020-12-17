package main

import (
    "log"
    "os"
)

type Ofst int64
type Amnt int64

const (
    numIter = 7
    blkStartHeight = 0  //it is fixed now! do not change
    blkHeight = BlockLimit
)


var (
    loggerI *log.Logger
    loggerE *log.Logger
    loggerD *log.Logger
)

func main() {
    var i int32

    var Num_zero_mixin int64 = 0
    var Num_traced_txin int64 = 0
    var Num_total_txin int64 = 0

    //var RingCTSpent     map[Ofst]bool           = make(map[Ofst]bool)
    var NonRingCTSpent  map[Amnt]map[Ofst]bool  = make(map[Amnt]map[Ofst]bool)

    // --- log setting
    f, err := os.OpenFile("./log/phase1_v1.1.log",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0644)
    if err!=nil{
        log.Fatal(err)
    }
    defer f.Close()

    loggerI = log.New(f, "[INFO] ", log.LstdFlags|log.Lshortfile)
    loggerE = log.New(f, "[ERROR] ", log.LstdFlags|log.Lshortfile)
    loggerD = log.New(f, "[DEBUG] ", log.LstdFlags|log.Lshortfile)

    // --- db init 
    tb := NewTracingBlocks()
    defer tb.db.Close()

    tb.DBInit(blkHeight)

    // --- tracing inputs
    for iterBC:=0; iterBC<numIter; iterBC++ {
        loggerI.Printf("%d'th iteration.\n", iterBC)

        for i=blkStartHeight ; i<blkHeight ; i++ {
            block := tb.GetBlock(i)

            for _, ti := range block.TxInputs {

                if ti.Version == 1 {
                    if len(ti.Amounts) != len(ti.Goffsetss) {
                        loggerE.Println("len of amounts and goffsetss are different")
                    }

                    for j:=0; j<len(ti.Amounts); j++ { //each txin_v
                        var untraced_offsets []Ofst = make([]Ofst, 0, len(ti.Amounts))
                        var amnt Amnt = Amnt(ti.Amounts[j])

                        if iterBC==0 {
                            if Num_total_txin++; len(ti.Goffsetss[j])==1 {
                                Num_zero_mixin++
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
                            Num_traced_txin++
                        }
                    }
                } else if ti.Version == 2 {   //pass at this time
                } else {
                    loggerD.Println("other transaction version exist")
                }
            }

        }// end one blockchain
    } // end iterBC

    loggerI.Println("** Program Completed **")
    loggerI.Println("# of total txins :",Num_total_txin)
    loggerI.Println("# of zero mix-ins :",Num_zero_mixin)
    loggerI.Println("# of total traced txins (effective 0 mix-in) :",Num_traced_txin)

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
