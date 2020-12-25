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
    dbFile = "./dbfile/traceXMR.db"
    traceBucket = "inputs"
    BlockLimit  = 2238270
)

type TxInfo struct {   // info for each transaction
    Version     int64       // nonRingCT:0 & RingCT:1
    TxHash      []byte
    Amounts     []int64
    Goffsetss   [][]int64   // set of global offsets Vin:Offset
}

type BlockTxs struct {
    TxInputs    []*TxInfo
}
// NewBlockTxs, Serialization, DeserializeBlockTxs,

type TracingBlocks struct {
    db      *bolt.DB
    length  int32
}
// NewTracingBlocks, PutBlock, GetBlock, DBInit,
// TimestampInit, PutTimestamp, GetTimestamp

//---
func NewBlockTxs(tis []*TxInfo) *BlockTxs {
    bt := new(BlockTxs)
    bt.TxInputs = tis
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
            b, err := tx.CreateBucket([]byte(traceBucket))
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

            //for timestamp
            ltvalue := b.Get([]byte("lt"))
            if ltvalue!=nil {
                length, err = strconv.Atoi(string(ltvalue))
                if err!=nil {
                    return err
                }
                Timelen = int32(length)
            } else {
                Timelen = 0
                err = b.Put([]byte("lt"),[]byte("0"))
                if err!=nil {
                    return err
                }
            }
        }

        return nil
    })
    if err!=nil {
        loggerE.Println("db creation failed")
        os.Exit(1)
    }

    tb.DBInit(BlockLimit)
    tb.TimestampInit(BlockLimit)

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
        os.Exit(1)
    }
}

func (tb *TracingBlocks) GetBlock (i int32) *BlockTxs {
    var v []byte
    if i >= tb.length {
        loggerE.Println("out of index: %d >= %d\n", i, tb.length)
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

        txHashes := NCBTxsFromBlock(i)
        txInfos := GetTxInputInfo(txHashes)
        tb.PutBlock(NewBlockTxs(txInfos))
    }

    loggerI.Printf("** db update finished **\n")
}

var Timelen int32

func (tb *TracingBlocks) PutTimestamp (timestamp []byte) {
    err := tb.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(traceBucket))
        index := strconv.FormatInt(int64(Timelen),10)
        index += "t"

        err := b.Put([]byte(index), timestamp)
        if err!=nil {
            return err
        }

        Timelen++
        err = b.Put([]byte("lt"),[]byte(fmt.Sprint(Timelen)))
        if err!=nil {
            return err
        }

        return nil
    })
    if err!=nil {
        loggerE.Println(err)
        os.Exit(1)
    }
}

func (tb *TracingBlocks) GetTimestamp (i int32) []byte {
    var v []byte
    if i>=Timelen {
        loggerE.Println("out of index: %d >= %d\n", i, Timelen)
        return nil
    }
    err := tb.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(traceBucket))
        index := fmt.Sprint(i) + "t"
        v = b.Get([]byte(index))
        return nil
    })
    if err!=nil {
        loggerE.Println(err)
    }
    return v
}

func (tb *TracingBlocks) TimestampInit(height int32) {
    if Timelen == height {
        loggerI.Printf("db timestamp is fully synchronized (length: %d)...\n", Timelen)
        return
    }

    loggerI.Printf("** db timestamp update start with length: %d **\n", Timelen)

    var apercent int32 = height/int32(100)
    var progress int32 = Timelen/apercent

    for i:=Timelen; i<height; i++ {
        if progress == i/apercent {
            loggerI.Printf("db update progress: %d%%...\n", progress)
            progress++
        }

        timestamp := GetBlockTimestamp(i)
        tb.PutTimestamp(timestamp)
    }

    loggerI.Printf("** db update finished **\n")
}

