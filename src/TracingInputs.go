package main

// ---
type TracingTxInput struct {        // for each transactions
    TxInputs        *TxInputInfo    // {Version, TxHash, Amounts, Goffsetss}
}

func NewTracingTxInput(tii *TxInputInfo) *TracingTxInput {
    tti := new(TracingTxInput)
    tti.TxInputs = tii
    return tti
}

// --- 
type TracingInputsBlock struct {
    TracingTxInputs [][]*TracingTxInput //each txs per block
    length          int32
}

func NewTracingInputsBlock(length int32) *TracingInputsBlock {
    tib := new(TracingInputsBlock)
    tib.TracingTxInputs = make([][]*TracingTxInput,length)
    tib.length = 0
    return tib
}

func (tib *TracingInputsBlock) AddTxsForBlock(tiis []*TxInputInfo, capacity int32) {
    tib.TracingTxInputs[tib.length] = make([]*TracingTxInput,len(tiis))
    for i, tii := range tiis {
        tib.TracingTxInputs[tib.length][i] = NewTracingTxInput(tii)
    }
    if len(tib.TracingTxInputs[tib.length]) > 0 {
        loggerD.Printf("%d block, %d txs\n", tib.length, len(tib.TracingTxInputs[tib.length]))
    }
    tib.length = (tib.length+1) % capacity
}
