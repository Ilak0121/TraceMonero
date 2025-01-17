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
    logVer = "v1.10"
    logFile = "./log/tracelog_"+logVer+".log"
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

    phase2(tb)

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
