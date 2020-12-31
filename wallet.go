package coinselection

import (
	"errors"
	"math/rand"
)

type Wallet interface {
	// CreateTx must use the wallets UTXO set to construct a transaction that fulfills target output UTXO given the fee rate.
	// It should return the input UTXOs that it plans to use along with a set of optional change output UTXOs.
	CreateTx(output UTXO, feePerByte int) (inputIDs []string, change []UTXO, err error)

	// Receive must add the given utxo to the wallets utxo set.
	Receive(utxo UTXO) error

	// Remove must remove the utxo with the given ID from the wallets utxo set.
	Remove(ID string) error

	// GetUtxos should return the wallets current utxo set
	GetUtxos() []UTXO
}

func SpendFromWallet(w Wallet, output UTXO, feePerByte int) error {
	inputs, change, err := w.CreateTx(output, feePerByte)
	if err != nil {
		return err
	}

	err = ExecuteTx(w, inputs, change)
	if err != nil {
		return err
	}

	return nil
}

// ExecuteTx removes the tx inputs from the wallet and adds the new change output.
func ExecuteTx(w Wallet, inputIDs []string, change []UTXO) error {
	for _, id := range inputIDs {
		err := w.Remove(id)
		if err != nil {
			return err
		}
	}

	for _, u := range change {
		err := w.Receive(u)
		if err != nil {
			return err
		}
	}

	return nil
}

func WalletTotalBalance(w Wallet) int {
	var total int
	for _, u := range w.GetUtxos() {
		total += u.value
	}
	return total
}

// EffectiveBalance is the sum of all the utxo values minus the fee required to spend them
func EffectiveBalance(utxos []UTXO, feePerByte int) int {
	var total int
	for _, u := range utxos {
		total += u.EffectiveValue(feePerByte)
	}
	return total
}

type UTXO struct {
	ID         string
	value      int
	scriptType ScriptType
	// locktime   int // maybe later
}

func (u UTXO) EffectiveValue(feePerByte int) int {
	return u.value - (u.scriptType.InputBytes() * feePerByte)
}

type utxoParams struct {
	inputBytes  int
	outputBytes int
}

type ScriptType int

const (
	p2pkh      ScriptType = 0
	p2sh       ScriptType = 1
	p2wsh      ScriptType = 2
	p2wpkh     ScriptType = 3
	p2shP2wpkh ScriptType = 4
	p2shP2wsh  ScriptType = 5
)

var scriptTypes = map[ScriptType]utxoParams{
	p2pkh: {
		inputBytes:  148,
		outputBytes: 34,
	},
}

func (st ScriptType) InputBytes() int {
	return scriptTypes[st].inputBytes
}

func (st ScriptType) OutputBytes() int {
	return scriptTypes[st].outputBytes
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// genID generates a random 10 char ID.
func genID() string {
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func probabilisticTypeProvider(distribution map[ScriptType]int) (func() ScriptType, error) {
	var (
		types []ScriptType
		cdf   []int
		total int
	)

	for k, v := range distribution {
		types = append(types, k)
		total += v
		cdf = append(cdf, total)
	}

	if total != 100 {
		return nil, errors.New("probabilities should sum up to 1")
	}

	return func() ScriptType {
		r := rand.Intn(100)
		var st ScriptType

		for i := len(cdf) - 1; i >= 1; i-- {
			if r >= cdf[i-1] {
				st = types[i]
				break
			}
		}
		return st
	}, nil
}
