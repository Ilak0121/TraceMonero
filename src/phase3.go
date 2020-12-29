package main

const (
    totaltxin = 18427996
)

func phase3(tb *TracingBlocks) {

    count := 0
    tp := int32(0)
    for i:=int32(blkStartHeight); i<blkHeight; i++ {
        block := tb.GetBlock(i)

        for _, ti := range block.TxInputs {
            if ti.IsCoinbase ==true || ti.Version ==2 {
                continue
            }
            for j, r_ofst := range ti.Roffsets {
                count++
                if r_ofst!=0{
                    if i_ofst := ti.Goffsetss[j][len(ti.Goffsetss[j])-1]; i_ofst==r_ofst {
                        tp++
                    }
                }
            }

        }
    }

    loggerI.Println("** Phase3 Completed **")
    loggerI.Printf("# of total txin : %d\n",    count)
    loggerI.Printf("# of total tp : %.2f\n",    float64(tp)/float64(count))

}
