package main

type Ofst int64
type Amnt int64

const (
    blkStartHeight = 0  //it is fixed now! do not change
    blkHeight = BlockLimit
)

func Phase1(numIter int, tb *TracingBlocks) (total_txin int64, traced_txin int64, zero_mixin int64, ) {
    zero_mixin = 0
    traced_txin = 0
    total_txin = 0

    //var RingCTSpent     map[Ofst]bool           = make(map[Ofst]bool)
    var NonRingCTSpent  map[Amnt]map[Ofst]bool  = make(map[Amnt]map[Ofst]bool)

    var i int32

    // --- tracing inputs
    for iterBC:=0; iterBC<numIter; iterBC++ {
        loggerI.Printf("%d'th iteration.\n", iterBC)

        for i=blkStartHeight ; i<blkHeight ; i++ {
            block := tb.GetBlock(i)

            for _, ti := range block.TxInputs {

                if ti.Version == 1 {
                    if len(ti.Amounts) != len(ti.Goffsetss) {
                        loggerE.Println("len of amounts and goffsetss are different")
                    }

                    for j:=0; j<len(ti.Amounts); j++ { //each txin_v
                        var untraced_offsets []Ofst = make([]Ofst, 0, len(ti.Amounts))
                        var amnt Amnt = Amnt(ti.Amounts[j])

                        if iterBC==0 {
                            if total_txin++; len(ti.Goffsetss[j])==1 {
                                zero_mixin++
                            }
                        }

                        for _, offset_r := range ti.Goffsetss[j] {
                            offset := Ofst(offset_r)
                            if _, ok := NonRingCTSpent[amnt][offset]; !ok{ //seen?
                                untraced_offsets = append(untraced_offsets, offset)
                            }
                        }

                        if len(untraced_offsets) == 1 {

                            if _, ok := NonRingCTSpent[amnt]; !ok{
                                NonRingCTSpent[amnt] = make(map[Ofst]bool)
                            }
                            NonRingCTSpent[amnt][untraced_offsets[0]] = true
                            traced_txin++
                        }
                    }
                } else if ti.Version == 2 {   //pass at this time
                } else {
                    loggerD.Println("other transaction version exist")
                }
            }

        }// end one blockchain
    } // end iterBC

    return
}
