package coinselection

import (
	"errors"
	"sort"
)

type GreedyWallet struct {
	utxos map[string]UTXO
}

func NewGreedyWallet(utxos []UTXO) (Wallet, error) {
	gw := GreedyWallet{
		utxos: map[string]UTXO{},
	}

	for _, u := range utxos {
		err := gw.Receive(u)
		if err != nil {
			return GreedyWallet{}, err
		}
	}

	return gw, nil
}

func (g GreedyWallet) CreateTx(output UTXO, feePerByte int) (inputs []string, change []UTXO, err error) {
	utxos := g.GetUtxos()
	sort.Slice(utxos, func(i, j int) bool {
		return utxos[i].EffectiveValue(feePerByte) > utxos[j].EffectiveValue(feePerByte)
	})

	outCost := output.scriptType.OutputBytes() * feePerByte

	var (
		totalIn     int
		effectiveIn int
	)

	for len(utxos) > 0 {
		totalIn += utxos[0].value
		effectiveIn += utxos[0].EffectiveValue(feePerByte)
		inputs = append(inputs, utxos[0].ID)
		if effectiveIn >= outCost+output.value {
			break
		}
		utxos = utxos[1:]
	}

	if effectiveIn < outCost+output.value {
		return nil, nil, ErrInsufficientBalance
	}

	// Check if having change makes sense:
	c := UTXO{
		ID:         genID(),
		value:      totalIn - output.value,
		scriptType: p2pkh,
	}
	outCost += c.scriptType.OutputBytes() * feePerByte

	if c.EffectiveValue(feePerByte) <= 0 || effectiveIn < outCost+output.value {
		return inputs, nil, nil
	}

	return inputs, []UTXO{c}, nil
}

func (g GreedyWallet) Receive(utxo UTXO) error {
	if _, ok := g.utxos[utxo.ID]; ok {
		return errors.New("utxo already exists in wallet")
	}

	g.utxos[utxo.ID] = utxo
	return nil
}

func (g GreedyWallet) Remove(ID string) error {
	if _, ok := g.utxos[ID]; !ok {
		return errors.New("utxo does not exists in wallet")
	}

	delete(g.utxos, ID)
	return nil
}

func (g GreedyWallet) GetUtxos() []UTXO {
	var res []UTXO
	for _, u := range g.utxos {
		res = append(res, u)
	}

	return res
}

var _ Wallet = (*GreedyWallet)(nil)
