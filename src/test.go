package main

import (
    //"encoding/json"
    "fmt"
    //"reflect"
    "time"
    "log"
    "net/http"
    //"net/url"
    "os"
    "math"
    "io/ioutil"
    "bytes"
    //"strconv"
    "github.com/buger/jsonparser"
)

const (
    urlTx       = "http://127.0.0.1:18081/get_transactions"
    urlRPC      = "http://127.0.0.1:18081/json_rpc"
    blockLimit  = 1000//2238270
)

/*
    GetTx (txHash []byte)
    txHash = list of transaction hashes. e.g. ['123124513412','1253523523']
    return: body contents as byte string, i.e. []byte
*/
func GetTx(txHash []byte) []byte{

    str := fmt.Sprintf(`decode_as_json:True, "tx_hashes":%s`, string(txHash))
    jsonStr := []byte(str)

    req, err := http.NewRequest("POST",urlTx,bytes.NewBuffer(jsonStr))
    req.Header.Set("content-type","application/json")

    client := &http.Client{}
    client.Timeout = time.Second * 15

    resp, err := client.Do(req)
    if err != nil{
        log.Fatal(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode!=200 {
        log.Fatal("[ERROR] response not 200")
    }

    body,_ := ioutil.ReadAll(resp.Body)
    
    return body
}

/*
    GetBlock (block_height int32)

*/
func GetBlock(block_height int32) []byte {
    str := fmt.Sprintf(`{"method":"get_block", "jsonrpc":"2.0", "id":"0", "params":{"height":%d}`, block_height)
    jsonStr := []byte(str)

    req, err := http.NewRequest("POST",urlRPC,bytes.NewBuffer(jsonStr))
    req.Header.Set("content-type","application/json")

    client := &http.Client{}
    client.Timeout = time.Second * 15

    resp, err := client.Do(req)
    if err != nil{ // is this containing 404-like?
        log.Fatal(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode!=200 {
        log.Fatal("[ERROR] response not 200")
    }

    body,_ := ioutil.ReadAll(resp.Body)
    
    return body
}

func main(){
    var i int32
    var body []byte
    f, err := os.OpenFile("text.log",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0644)
    if err!=nil{
        log.Fatal(err)
    }
    defer f.Close()
    loggerI := log.New(f, "[INFO]", log.LstdFlags)
    loggerE := log.New(f, "[ERROR]", log.LstdFlags)
    
    CBTxNum := 0
    NCBTxNum := 0

    start := time.Now()
    
    var percentIndex int32 = 0
    for i=0;i<blockLimit+1;i++{
        body = GetBlock(i)

        //minerTxHash
        _, err := jsonparser.GetString(body,"result","miner_tx_hash")
        if err!=nil{
            loggerE.Println(err)
        }
        CBTxNum += 1 //miner tx count

        //fmt.Println("[INFO] minerTxHash:",minerTxHash)

        _, err = jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error){
            if err!=nil{
                loggerE.Println(err)
            }
            NCBTxNum += 1 //txs count
            //fmt.Println("[INFO]",i,"'th", "minerTxHash:", string(value))
        },"result","tx_hashes")
        // ignore the err of non existing case of non-coinbase transactions
        percentI := int32(math.Trunc(100*float64(i)/float64(blockLimit)))
        percentS := fmt.Sprintf("%d",percentI)
        if percentIndex == percentI{
            loggerI.Println(percentS,"...")
            percentIndex += 1
        }
    }

    duration := time.Since(start)

    loggerI.Println("[INFO] Done")
    loggerI.Println("[INFO] # of Coinbase Tx",CBTxNum)
    loggerI.Println("[INFO] # of non-Coinbase Tx",NCBTxNum)
    loggerI.Println("[INFO] elapsed time:",duration)

}
