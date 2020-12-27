package main

const (
    totaltxin = 18427996
    totaltx = 18427996/10 //?
)

type Pair struct {
    Amnt    int64
    Ofst    int64
}

func phase2(tb *TracingBlocks) {
    h2 := make([][]string,totaltx) //destined tx -> source txs
    oi := tb.GetOutInfo()

    for i:=int32(blkStartHeight); i<blkHeight; i++{
        block := tb.GetBlock(i)

        for j, ti := range block.TxInputs {
            var list []string
            tmp_t := make(map[string]bool)

            for k, ofsts := range ti.Goffsetss { //vins
                tmp := make(map[string]bool)

                for _, ofst := range ofsts { //ofsts each vin
                    hash, _, _, err := oi.GetInfo(ti.Amounts[k], ofst, tb)
                    if err!=nil {
                        loggerE.Println(err)
                    }
                    tmp[string(hash)] = true
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

    var tp, fp, up int64 = 0, 0, 0
    for i:=int32(blkStartHeight); i<blkHeight; i++{ //comparing with true positive
        block := tb.GetBlock(i)

        for j, ti := range block.TxInputs {

            tmp_r   := make(map[string]bool)
            tmp_map := make(map[string]bool)
            for k, ofst := range ti.Roffsets {
                hash, _, _, err := oi.GetInfo(ti.Amounts[k], ofst, tb)
                if err!=nil{
                    loggerE.Println(err)
                }
                if _, ok := tmp_map[string(hash)]; !ok{
                    tmp_map[string(hash)] = true
                } else {
                    tmp_r[string(hash)] = true
                }
            }

            h2_map := make(map[string]bool)
            for _, v := range h2[j] {
                h2_map[v]=true
            }

            flagn := 0
            for k, _ := range tmp_map {
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
