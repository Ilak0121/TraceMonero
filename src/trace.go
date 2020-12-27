package main

import (
    "os"
    "fmt"
    "bytes"
    "errors"
    "strconv"
    "encoding/gob"
    "github.com/boltdb/bolt"
)

const (
    dbFile = "./dbfile/traceXMR2.db"
    traceBucket = "inputs"
    BlockLimit  = 2238270
    BlockHeightofPaper = 1240000
)

type TxInfo struct {        // info for each transaction
    IsCoinbase  bool    //**
    Version     int64       // nonRingCT:1 & RingCT:2
    TxHash      []byte

    Amounts     []int64
    Goffsetss   [][]int64   // set of global offsets Vin:Offset
    Roffsets    []int64 //**

    OutIndices  []int64
    OutAmounts  []int64
}

type BlockTxs struct {
    TxInputs    []*TxInfo
    Timestamp   []byte
}
// NewBlockTxs, Serialization, DeserializeBlockTxs, GetTimestamp

type OutInfo struct {
    OfstToHash      map[int64]map[int64][]byte  //txHash
    THtoHeight      map[string]int32
}
// GetInfo - Hash, Height, Time 
// SetInfo - Hash, Height, 
// Serialization, DeserializeOutInfo

type TracingBlocks struct {
    db      *bolt.DB
    length  int32
}
// NewTracingBlocks, PutBlock, GetBlock, UpdateBlock, DBInit,
// PutOutInfo, GetOutInfo, OutInfoInit

//---
func NewBlockTxs(tis []*TxInfo, timestamp []byte) *BlockTxs {
    bt := new(BlockTxs)
    bt.TxInputs = tis
    bt.Timestamp = timestamp
    return bt
}

func (bt *BlockTxs) Serialization() []byte {
    var result bytes.Buffer

    encoder := gob.NewEncoder(&result)

    err := encoder. Encode(bt)
    if err!=nil{
        loggerE.Println(err)
    }

    return result.Bytes()
}

func DeserializeBlockTxs (d []byte) *BlockTxs {
    var bt BlockTxs

    decoder := gob.NewDecoder(bytes.NewReader(d))

    err := decoder.Decode(&bt)
    if err!=nil {
        loggerE.Println(err)
    }

    return &bt
}

func (bt *BlockTxs) GetTimestamp() []byte {
    return bt.Timestamp
}

// ---
func (oi *OutInfo) GetInfo(amnt int64, ofst int64, tb *TracingBlocks) ([]byte, int32, []byte, error) {
    var hash, timestamp []byte
    var height          int32
    var ok              bool

    hash, ok = oi.OfstToHash[amnt][ofst]
    height, ok = oi.THtoHeight[string(hash)]

    if ok==false {
        return nil, 0, nil, errors.New("key does not exist")
    }
    timestamp = tb.GetBlock(height).Timestamp

    return hash, height, timestamp, nil
}

func (oi *OutInfo) SetInfo(amnt int64, ofst int64, hash []byte, height int32) {
    if _, ok := oi.OfstToHash[amnt]; !ok {
        oi.OfstToHash[amnt] = make(map[int64][]byte)
    }

    oi.OfstToHash[amnt][ofst] = hash
    oi.THtoHeight[string(hash)] = height
}

func (oi *OutInfo) Serialization() []byte {
    var result bytes.Buffer

    encoder := gob.NewEncoder(&result)

    err := encoder.Encode(oi)
    if err!=nil{
        loggerE.Println(err)
    }

    return result.Bytes()
}

func DeserializeOutInfo(d []byte) *OutInfo {
    var oi OutInfo

    decoder := gob.NewDecoder(bytes.NewReader(d))

    err := decoder.Decode(&oi)
    if err!=nil {
        loggerE.Println(err)
    }

    return &oi
}

//---
func NewTracingBlocks() *TracingBlocks {
    tb := new(TracingBlocks)

    db,err := bolt.Open(dbFile,0600,nil)
    if err!=nil {
        loggerE.Println(err)
        os.Exit(1)
    }
    tb.db = db

    err = db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(traceBucket))
        if b==nil {
            b,err := tx.CreateBucket([]byte(traceBucket))
            if err!=nil {
                return err
            }

            err = b.Put([]byte("l"),[]byte("0"))
            if err!=nil {
                return err
            }

            tb.length = 0
        } else {
            length, err := strconv.Atoi(string(b.Get([]byte("l"))))
            if err!=nil {
                return err
            }
            tb.length = int32(length)
        }

        return nil
    })
    if err!=nil {
        loggerE.Println(err)
        os.Exit(1)
    }
    tb.DBInit(BlockHeightofPaper)
    tb.OutInfoInit(BlockHeightofPaper)

    return tb
}

func (tb *TracingBlocks) PutBlock (bt *BlockTxs) {
    err := tb.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(traceBucket))
        index := strconv.FormatInt(int64(tb.length),10)

        err := b.Put([]byte(index),bt.Serialization())
        if err!=nil {
            return err
        }

        tb.length++
        err = b.Put([]byte("l"),[]byte(fmt.Sprint(tb.length)))
        if err!=nil {
            return err
        }

        return nil
    })
    if err!=nil {
        loggerE.Println(err)
    }
}

func (tb *TracingBlocks) GetBlock (i int32) *BlockTxs {
    var v []byte
    if i >= tb.length {
        loggerE.Println("out of index: %d\n",tb.length)
        return nil
    }
    err := tb.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(traceBucket))
        v = b.Get([]byte(fmt.Sprint(i)))
        return nil
    })
    if err!=nil {
        loggerE.Println(err)
    }
    return DeserializeBlockTxs(v)
}

func (tb *TracingBlocks) UpdateBlock (height int32, offset int, ti *TxInfo) {
    err := tb.db.Batch(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(traceBucket))
        index := strconv.FormatInt(int64(height),10)

        bt := DeserializeBlockTxs(b.Get([]byte(fmt.Sprint(height))))
        bt.TxInputs[offset] = ti

        err := b.Put([]byte(index),bt.Serialization())
        if err!=nil {
            return err
        }

        return nil
    })
    if err!=nil {
        loggerE.Println(err)
    }
}

func (tb *TracingBlocks) DBInit(height int32) {
    if tb.length == height {
        loggerI.Printf("db is fully synchronized (length: %d)...\n", tb.length)
        return
    }

    loggerI.Printf("** db update start with length: %d **\n", tb.length)

    var apercent int32 = height/int32(100)
    var progress int32 = tb.length/apercent

    for i:=tb.length; i<height; i++ {
        if progress == i/apercent {
            loggerI.Printf("db update progress: %d%%...\n", progress)
            progress++
        }

        txHashes, timestamp := GetTxsFromBlock(i)
        txInfos := GetTxInputInfo(txHashes)
        tb.PutBlock(NewBlockTxs(txInfos, timestamp))
    }

    loggerI.Printf("** db update finished **\n")
}

func (tb *TracingBlocks) GetOutInfo() *OutInfo {
    var v   []byte

    _ = tb.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(traceBucket))
        v = b.Get([]byte("oi"))
        return nil
    })

    if v==nil {
        return nil
    } else {
        return DeserializeOutInfo(v)
    }
}

func (tb *TracingBlocks) PutOutInfo(oi *OutInfo) {
    err := tb.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(traceBucket))
        err := b.Put([]byte("oi"), oi.Serialization())
        if err!=nil {
            return err
        }
        return nil
    })
    if err!=nil {
        loggerE.Println(err)
    }
}

func (tb *TracingBlocks) OutInfoInit(height int32) {
    if tb.GetOutInfo() != nil {
        loggerI.Printf("OutInfo is fully synchronized (length: %d)...\n", tb.length)
        return
    }

    loggerI.Printf("** OutInfo update start... **\n")

    var tmp *OutInfo = new(OutInfo)
    tmp.OfstToHash = make(map[int64]map[int64][]byte)
    tmp.THtoHeight = make(map[string]int32)

    var apercent int32 = height/int32(100)
    var progress int32 = int32(0)

    for i:=int32(0); i<height; i++ {
        block := tb.GetBlock(i)
        for _, ti := range block.TxInputs {
            for j:=0; j<len(ti.OutIndices); j++ {
                tmp.SetInfo(ti.OutAmounts[j], ti.OutIndices[j], ti.TxHash, i)
            }
        }

        if progress == i/apercent {
            loggerI.Printf("OutInfo update progress: %d%%...\n", progress)
            progress++
        }
    }
    tb.PutOutInfo(tmp)

    loggerI.Printf("** OutInfo update finished **\n")
}

