package main

import (
    "fmt"
    "time"
    "net/http"
    "io/ioutil"
    "bytes"
    "strconv"
    "github.com/buger/jsonparser"
)

const (
    urlTx       = "http://127.0.0.1:18081/get_transactions"
    urlRPC      = "http://127.0.0.1:18081/json_rpc"
)

/*
* GetBlock
* GetTx
* NCBTxsFromBlock
* GetTxInputInfo
* GetBlockTimestamp
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
        loggerE.Println(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode!=200 {
        loggerE.Println("[ERROR] response not 200")
    }

    body,_ := ioutil.ReadAll(resp.Body)

    return body
}

/*
    @Brief: return response body of RPC request 'get_transaction' by given set of transaction hashes.
    input:
        txHash = list of transaction hashes - e.g. ['123124513412','1253523523']
    return: 
        body contents as []byte
*/
func GetTx(txHashes [][]byte) []byte {
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

/**
* @Brief: return transaction hashes from given block height.
*/
func NCBTxsFromBlock(block_height int32) [][]byte {
    var txHashes [][]byte

    body := GetBlock(block_height)

    jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error){
        txHashes = append(txHashes, value)
    },"result","tx_hashes")

    return txHashes
}

func GetTxInputInfo(txHashes [][]byte) ([]*TxInfo) {
    var txInfos []*TxInfo

    body := GetTx(txHashes)

    // for each tx
    jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error){
        var amounts     []int64
        var goffsetss   [][]int64

        txHash, _, _, _ := jsonparser.Get(value, "tx_hash")
        asJson, _, _, _ := jsonparser.Get(value, "as_json")
        version, _      := jsonparser.GetInt(asJson, "version")

        //for each txin_v, get amount and goffsets
        jsonparser.ArrayEach(asJson, func(value []byte, dataType jsonparser.ValueType, offset int, err error){
            var amount, base int64 = 0, 0
            var goffsets []int64

            amount, _ = jsonparser.GetInt(value, "key", "amount")

            jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error){
                buf, _ := strconv.Atoi(string(value))
                loffset := int64(buf)
                goffsets = append(goffsets, loffset+base)
                base += loffset
            },"key","key_offsets")

            amounts = append(amounts, amount)
            goffsetss = append(goffsetss, goffsets)

        },"vin")

        txInfos = append(txInfos, &TxInfo{version, txHash, amounts, goffsetss})

    }, "txs")

    return txInfos
}

func GetBlockTimestamp(i int32) []byte {
    body := GetBlock(i)

    timestamp, _, _, err := jsonparser.Get(body, "result", "block_header", "timestamp")
    if err!=nil {
        loggerE.Println(err)
    }

    time.Sleep(1000 * time.Millisecond) //to lower the burden
    return timestamp
}

/* legacy code
func GetTxData(txHashes [][]byte) (jsons []string, indices []string){
    body := GetTx(txHashes)
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
}*/

