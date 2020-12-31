package coinselection

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestWalletBasics tests that your wallet isn't a complete dud and that its basics are in order.
func TestWalletBasics(t *testing.T) {
	wallets := []struct {
		name     string
		provider func(utxos []UTXO) (Wallet, error)
	}{
		{
			name:     "Greedy",
			provider: NewGreedyWallet,
		},
	}

	for _, w := range wallets {
		t.Run(w.name, func(t *testing.T) {
			gw, err := w.provider([]UTXO{utxo1})
			require.NoError(t, err)
			require.Equal(t, 100000000, WalletTotalBalance(gw))
			require.Equal(t, 99999852, EffectiveBalance(gw.GetUtxos(), 1))

			err = gw.Receive(utxo2)
			require.NoError(t, err)
			require.Equal(t, 250000000, WalletTotalBalance(gw))
			require.Equal(t, 249970400, EffectiveBalance(gw.GetUtxos(), 100))

			err = ExecuteTx(gw, []string{utxo1.ID}, []UTXO{utxo3})
			require.NoError(t, err)
			require.Equal(t, 150050000, WalletTotalBalance(gw))
			require.Equal(t, 150047040, EffectiveBalance(gw.GetUtxos(), 10))

			err = ExecuteTx(gw, []string{utxo2.ID, utxo3.ID}, nil)
			require.NoError(t, err)
			require.Equal(t, 0, WalletTotalBalance(gw))
			require.Equal(t, 0, EffectiveBalance(gw.GetUtxos(), 10))

			err = gw.Receive(utxo1)
			require.NoError(t, err)

			err = SpendFromWallet(gw, utxo1, 0)
			require.NoError(t, err)
			require.Equal(t, 0, WalletTotalBalance(gw))

			err = SpendFromWallet(gw, utxo1, 0)
			require.Error(t, err)
		})
	}
}

type result struct {
	paymentsCompleted         int
	effectiveRemainingBalance int
}

type results map[string]result

func TestWallets(t *testing.T) {
	wallets := []struct {
		name     string
		provider func(utxos []UTXO) (Wallet, error)
	}{
		{
			name:     "Greedy",
			provider: NewGreedyWallet,
		},
	}

	rand.Seed(time.Now().Unix())

	typeProvider, err := probabilisticTypeProvider(map[ScriptType]int{
		p2pkh: 100,
	})
	require.NoError(t, err)

	utxoBag := genUtxoSet(100, typeProvider, 100000, 100000000)
	paymentsBag := genUtxoSet(100, typeProvider, 1000, 500000000)
	feeRate := 10

	res := make(results)

	for _, w := range wallets {
		t.Run(w.name, func(t *testing.T) {
			wallet, err := w.provider(utxoBag)
			require.NoError(t, err)

			var numComplete int

			for _, p := range paymentsBag {
				err := SpendFromWallet(wallet, p, feeRate)
				if errors.Is(err, ErrInsufficientBalance) {
					break
				}
				require.NoError(t, err)

				numComplete++
			}

			res[w.name] = result{
				paymentsCompleted:         numComplete,
				effectiveRemainingBalance: EffectiveBalance(wallet.GetUtxos(), feeRate),
			}

		})
	}

	fmt.Println(res)
}

var (
	utxo1 = UTXO{
		ID:         "1",
		value:      100000000,
		scriptType: p2pkh,
	}
	utxo2 = UTXO{
		ID:         "2",
		value:      150000000,
		scriptType: p2pkh,
	}
	utxo3 = UTXO{
		ID:         "3",
		value:      50000,
		scriptType: p2pkh,
	}
	utxo4 = UTXO{
		ID:         "4",
		value:      10000,
		scriptType: p2pkh,
	}
)

func TestTypeProvider(t *testing.T) {
	dm := map[ScriptType]int{
		p2pkh:  20,
		p2sh:   60,
		p2wsh:  10,
		p2wpkh: 10,
	}

	rand.Seed(time.Now().Unix())

	provider, err := probabilisticTypeProvider(dm)
	require.NoError(t, err)

	counts := map[ScriptType]int{}

	for i := 0; i < 10000; i++ {
		counts[provider()]++
	}

	fmt.Println(counts)

	// Cant test a probabilistic function. Just run it a few times look at results until you are
	// convinced that it works.
}

// TODO(elle): Maybe add a value distribution instead of min, max
func genUtxoSet(n int, typeProvider func() ScriptType, min, max int) []UTXO {
	var set []UTXO
	for i := 0; i < n; i++ {
		set = append(set, UTXO{
			ID:         genID(),
			value:      min + rand.Intn(max-min),
			scriptType: typeProvider(),
		})
	}

	return set
}
