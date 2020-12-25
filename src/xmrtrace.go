package main

import (
    "log"
    "os"
    "fmt"
    "strings"
    "encoding/csv"
)

var (
    loggerI *log.Logger
    loggerE *log.Logger
    loggerD *log.Logger
)

const (
    logVer = "v1.3"
    logFile = "./log/phase1_"+logVer+".log"
    TotalInputsFile = "./data/totalInputsperBlk_"+logVer+".csv"
    TotalTracedInputsFile = "./data/totalTracedInputsperBlk_"+logVer+".csv"
)

func main() {

    // --- log setting
    f, err := os.OpenFile(logFile,os.O_APPEND|os.O_CREATE|os.O_WRONLY,0644)
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

    //Phase1(tb)

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

func CSVWrite(data []int, file string) error {
    f, err := os.OpenFile(file,os.O_APPEND|os.O_CREATE|os.O_WRONLY,0644)
    if err!=nil {
        return err
    }
    defer f.Close()
    w := csv.NewWriter(f)
    defer w.Flush()

    buf := []string(strings.Fields(strings.Trim(fmt.Sprint(data),"[]")))
    if err=w.Write(buf); err!=nil {
        return err
    }
    return nil
}
