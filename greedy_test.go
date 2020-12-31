package coinselection

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGreedyWallet(t *testing.T) {
	gw, err := NewGreedyWallet([]UTXO{utxo1, utxo2, utxo3})
	require.NoError(t, err)

	inputs, change, err := gw.CreateTx(utxo4, 100)
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	require.Len(t, change, 1)
	require.Equal(t, "2", inputs[0])
	require.Equal(t, 149990000, change[0].value)

	gw, err = NewGreedyWallet([]UTXO{utxo1, utxo2, utxo3})
	require.NoError(t, err)

	err = SpendFromWallet(gw, utxo4, 100)
	require.NoError(t, err)
	require.Equal(t, 250040000, WalletTotalBalance(gw))
}
