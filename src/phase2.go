package main

import (
)

const (
    totaltx = 2199836
)

func phase2(tb *TracingBlocks) {
    h2 := make([][]string,totaltx) //destined tx -> source txs
    oi := tb.GetOutInfo()

    loggerI.Println("h2 start")

    apercent := blkHeight/int32(100)
    progress := int32(0)

    for i:=int32(blkStartHeight); i<blkHeight; i++{
        if progress == i/apercent {
            loggerI.Printf("one iteration progress: %d%%...\n", progress)
            progress++
        }

        block := tb.GetBlock(i)

        for j, ti := range block.TxInputs {
            if ti.IsCoinbase ==true || ti.Version==2 {
                continue
            }
            var list []string
            tmp_t := make(map[string]bool,len(ti.Goffsetss)*(len(ti.Goffsetss[0])+10))

            for k, ofsts := range ti.Goffsetss { //vins
                tmp := make(map[string]bool, len(ofsts))

                for _, ofst := range ofsts { //ofsts each vin
                    hash, _, _, err := oi.GetInfo(Pair{ti.Amounts[k], ofst}, tb)
                    if err==nil {
                        //loggerE.Println(err)
                        //loggerE.Printf("%v\n",Pair{ti.Amounts[k], ofst})
                        tmp[string(hash)] = true
                    }
                }

                for hash, _ := range tmp {
                    if _, ok := tmp_t[hash]; !ok {
                        tmp_t[hash]=true
                    } else {
                        list = append(list, hash)
                    }
                }

            }
            h2[j] = append(h2[j], list...)
        }//each tx
    }//each block

    loggerI.Println("TP start")

    var tp, fp, up int64 = 0, 0, 0

    progress = int32(0)
    for i:=int32(blkStartHeight); i<blkHeight; i++{ //comparing with true positive
        if progress == i/apercent {
            loggerI.Printf("one iteration progress: %d%%...\n", progress)
            progress++
        }

        block := tb.GetBlock(i)

        for j, ti := range block.TxInputs {
            if ti.IsCoinbase ==true || ti.Version==2{
                continue
            }
            var tmp_r []string
            tmp_map := make(map[string]bool,len(ti.Goffsetss)*(len(ti.Goffsetss[0])+10))

            for k, ofst := range ti.Roffsets {
                hash, _, _, err := oi.GetInfo(Pair{ti.Amounts[k], ofst}, tb)

                if err==nil{
                    if _, ok := tmp_map[string(hash)]; !ok{
                        tmp_map[string(hash)] = true
                    } else {
                        tmp_r = append(tmp_r, string(hash))
                    }
                }

            }

            h2_map := make(map[string]bool, len(h2[j]))
            for _, v := range h2[j] {
                h2_map[v]=true
            }

            flagn := 0
            for _, k := range tmp_r {
                if _, ok := h2_map[k]; ok {
                    flagn++
                }
            }

            if len(h2_map) == flagn {
                tp++
            } else if flagn == 0 {
                fp++
            } else {
                up++
            }
        }
    }
    loggerI.Println("** Phase2 Completed **")
    loggerI.Printf("# of total tx : %d\n",      totaltx)
    loggerI.Printf("# of total tp : %.2f\n",    float64(tp)/float64(totaltx))
    loggerI.Printf("# of total up : %.2f\n",    float64(up)/float64(totaltx))
    loggerI.Printf("# of total fp : %.2f\n",    float64(fp)/float64(totaltx))

}
