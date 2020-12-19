package main

import (
    "log"
    "os"
)

var (
    loggerI *log.Logger
    loggerE *log.Logger
    loggerD *log.Logger
)

func main() {

    // --- log setting
    f, err := os.OpenFile("./log/phase1_v_test.log",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0644)
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

    // numIter, *TracingBlocks
    total_txin, traced_txin, zero_mixin := Phase1(7, tb)

    loggerI.Println("** Program Completed **")
    loggerI.Println("# of total txins :",total_txin)
    loggerI.Println("# of zero mix-ins :",zero_mixin)
    loggerI.Println("# of total traced txins (effective 0 mix-in) :",traced_txin)

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
