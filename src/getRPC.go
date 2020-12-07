package main

import (
    "fmt"
    //"reflect"
    "time"
    "log"
    "net/http"
    //"net/url"
    //"os"
    //"math"
    "io/ioutil"
    "bytes"
    //"encoding/json"
    "strconv"
    "github.com/buger/jsonparser"
)

const (
    urlTx       = "http://127.0.0.1:18081/get_transactions"
    urlRPC      = "http://127.0.0.1:18081/json_rpc"
    BlockLimit  = 2238270

)

/*
    GetBlock (block_height int32)

*/
func GetBlock(block_height int32, loggerE *log.Logger) []byte {
    str := fmt.Sprintf(`{"method":"get_block", "jsonrpc":"2.0", "id":"0", "params":{"height":%d}`, block_height)
    jsonStr := []byte(str)

    req, err := http.NewRequest("POST",urlRPC,bytes.NewBuffer(jsonStr))
    req.Header.Set("content-type","application/json")

    client := &http.Client{}
    client.Timeout = time.Second * 15

    resp, err := client.Do(req)
    if err != nil{ // is this containing 404-like?
        loggerE.Println(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode!=200 {
        loggerE.Println("[ERROR] response not 200")
    }

    body,_ := ioutil.ReadAll(resp.Body)
    
    return body
}

/**
* @Brief: return transaction hashes from given block height.
*/
func NCBTxsFromBlock(block_height int32, loggerE *log.Logger) [][]byte {
    var txHashes [][]byte

    body := GetBlock(block_height, loggerE)

    jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error){
        txHashes = append(txHashes, value)
    },"result","tx_hashes")

    return txHashes
}

/*
    @Brief: return response body of RPC request 'get_transaction' by given set of transaction hashes.
    GetTx (txHash []string) []byte
    input:
        txHash = list of transaction hashes
        - e.g. ['123124513412','1253523523']
    return: 
        body contents as []byte
*/
func GetTx(txHashes [][]byte, loggerE *log.Logger) []byte {
    //qeury string
    str := string(`{"decode_as_json":True, "txs_hashes":[`)
    for i, txHash := range txHashes {
        data := string(txHash)
        if i != len(txHashes)-1 {
            str += fmt.Sprintf(`"%s",`,data)
        }else{
            str += fmt.Sprintf(`"%s"`,data)
        }
    }
    str += "]}"
    jsonStr := []byte(str)

    req, err := http.NewRequest("POST",urlTx,bytes.NewBuffer(jsonStr))
    req.Header.Set("content-type","application/json")

    client := &http.Client{}
    client.Timeout = time.Second * 15

    resp, err := client.Do(req)
    if err != nil{
        loggerE.Println(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode!=200 {
        loggerE.Println("[ERROR] response not 200")
    }

    body,_ := ioutil.ReadAll(resp.Body)

    return body
}

func GetTxData(txHashes [][]byte, loggerE *log.Logger) (jsons []string, indices []string){
    body := GetTx(txHashes, loggerE)
    jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error){
        asJson, err := jsonparser.GetString(value, "as_json")
        if err!=nil{
            loggerE.Println(err)
        }
        indice, _, _, err := jsonparser.Get(value, "output_indices")
        if err!=nil{
            loggerE.Println(err)
        }
        jsons = append(jsons, asJson)
        indices = append(indices, string(indice))
    }, "txs")

    return
}

/**
* @Brief: format of transaction input information; 
* Each vin contains multiple txin_v which includes mix_ins.
*/
type TxInputInfo struct {   // info for each transaction
    Version     int64       // nonRingCT:0 & RingCT:1
    TxHash      []byte
    Amounts     []int64
    Goffsetss   [][]int64   // set of global offsets
}

func GetTxInputInfo(txHashes [][]byte, loggerE *log.Logger) (txInfos []*TxInputInfo) {
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    body := GetTx(txHashes, loggerE)

    // for each tx
    jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error){
        var amounts     []int64
        var goffsetss   [][]int64

        txHash, _, _, err := jsonparser.Get(value, "tx_hash")
        if err!=nil {
            loggerE.Println(err)
        }

        asJson, err := jsonparser.GetString(value, "as_json")
        if err!=nil {
            loggerE.Println(err)
        }
        byteAsJson := []byte(asJson)

        //fmt.Println("[DBG] version:",string(asJson))
        version, err := jsonparser.GetInt(byteAsJson, "version")
        if err!=nil {
            loggerE.Println(err)
        }

        //for each txin_v, get amount and goffsets
        jsonparser.ArrayEach(byteAsJson, func(value []byte, dataType jsonparser.ValueType, offset int, err error){
            var amount int64
            var goffsets []int64

            amount, err = jsonparser.GetInt(value, "key", "amount")
            if err!=nil {
                loggerE.Println(err)
            }

            var base int64 = 0
            jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error){
                //fmt.Println("[DBG] value:",string(value))

                buf, err := strconv.Atoi(string(value))
                if err!=nil {
                    loggerE.Println(err)
                }
                loffset := int64(buf)
                goffsets = append(goffsets, loffset+base)
                base += loffset
            },"key","key_offsets")

            amounts = append(amounts, amount)
            goffsetss = append(goffsetss, goffsets)
        },"vin")

        txInfos = append(txInfos, &TxInputInfo{version, txHash, amounts, goffsetss})
    }, "txs")

    return
}
