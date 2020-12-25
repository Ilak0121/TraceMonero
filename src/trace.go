package main

import (
    "os"
    "fmt"
    "bytes"
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

type TracingBlocks struct {
    db      *bolt.DB
    length  int32
}

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

