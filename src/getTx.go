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
    //"strconv"
    "github.com/buger/jsonparser"
)

const (
    urlTx       = "http://127.0.0.1:18081/get_transactions"
    urlRPC      = "http://127.0.0.1:18081/json_rpc"
    blockLimit  = 2238270
)

/*
    GetTx (txHash []string) []byte
    input:
        txHash = list of transaction hashes
        - e.g. ['123124513412','1253523523']
    return: 
        body contents as []byte
*/
func GetTx(txHashes []string) []byte {
    //qeury string
    str := string(`{"decode_as_json":True, "txs_hashes":[`)
    for i, txHash := range txHashes {
        if i != len(txHashes)-1 {
            str += fmt.Sprintf(`"%s",`,txHash)
        }else{
            str += fmt.Sprintf(`"%s"`,txHash)
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

    log.SetFlags(log.LstdFlags | log.Lshortfile)

    var body []byte
    //body = GetTx([]string{"4953b8424b390a378502625c0ec2470668d3a16ae00787ec54cb161990957045","e0c627e5632e601eb733cc045b5210ee819bf1bbe3a846ada523c9397e359a96"})
    body = GetTx([]string{"4953b8424b390a378502625c0ec2470668d3a16ae00787ec54cb161990957045"})
    //fmt.Println("data:",string(body))

    _, err := jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error){

        asJson, err := jsonparser.GetString(value,"as_json")
        if err!=nil{
            log.Fatal(err)
        }
        fmt.Println("asJson:",asJson)

        _, err = jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error){
            fmt.Println("output_indices", string(value))
        },"output_indices")

    },"txs")
    if err!=nil {
        log.Fatal(err)
    }

    _, err = jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error){
        fmt.Println("txs_as_json:",string(value))
    },"txs_as_json")
    if err!=nil {
        log.Fatal(err)
    }

}
