package main

import (
    "fmt"
    "log"
    "strconv"
)


func main(){
    var blockheightstr string
    fmt.Scanln(&blockheightstr)
    parseHeight, err := strconv.ParseInt(blockheightstr, 10, 32)
    if err!=nil {
        log.Fatal(err)
    }

    var height int32 = int32(parseHeight)
    txs := NCBTxsFromBlock(height)
    for _,tx := range(txs) {
        fmt.Println("txH:", string(tx))
    }
    if len(txs) != 0 {
        txInputInfos := GetTxInputInfo(txs)
        for i, txinfo := range txInputInfos {
            fmt.Println(i, "th:")
            fmt.Println("version:", txinfo.Version)
            fmt.Println("txHash:", string(txinfo.TxHash))
            fmt.Println("amounts:", txinfo.Amounts)
            fmt.Println("goffsetss:", txinfo.Goffsetss)
        }
    }
}
/**
* Example result
2238270
txH: 4953b8424b390a378502625c0ec2470668d3a16ae00787ec54cb161990957045
0 th:
version: 2
txHash: 4953b8424b390a378502625c0ec2470668d3a16ae00787ec54cb161990957045
amounts: [0]
goffsetss: [[15472005 18413010 21279099 22372927 22513451 22870554 23404062 23611584 23627939 23652046 23656555]]
*/
