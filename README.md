# coinselection

Package `coinselection` provides a `Wallet` interface:

```
type Wallet interface {
	// CreateTx must use the wallets UTXO set to construct a transaction that fulfills target output UTXO given the fee rate. 
	// It should return the input UTXOs that it plans to use along with a set of optional change output UTXOs. 
	CreateTx(output UTXO, feePerByte int64) (inputs []UTXO, change []UTXO, err error)

	// Receive must add the given utxo to the wallets utxo set.
	Receive(utxo UTXO)

	// Remove must remove the utxo with the given ID from the wallets utxo set.
	Remove(ID string) error

	// GetUtxos should return the wallets current utxo set
	GetUtxos() []UTXO
}
```

A `Wallet` is made up of a collection of UTXOs (unspent Bitcoin outputs) and the main job of a `Wallet` is to 
construct transactions when the wallet owner wants to make a transaction. So, given an amount that the user wants to send, T,
the wallet must find the best inputs to use so that this amount can be satisfied. The wallet may result to making change: 
so for example if the wallet only has 1 utxo with value 10 and is requested to make a payment of 5, then it may construct a 
transaction with an input of 10 and 2 outputs, 1 to the recipient and 1 back to itself.

So far, this should be sounding a bit like the classic `Knapsack problem` or the `Subset sum problem`. But there are some more caveats:

- This is Bitcoin! So transaction fees are required. To spend any utxo, you will need to pay a price for spending it. And the same
goes for creating a utxo. So you need to think a bit about the tradeoffs of perhaps using many small utxos as input so that you dont need a change output 
but then paying the price for the size of all your inputs.
- UTXOs can be of different types. Each type of utxo will require a certain number of bytes to spend it (`input bytes`) and a certain number of bytes to create it (`output bytes`).

### Evaluation of different wallet implementations

The comparison of the different wallet implementations will be done as follows:

- A random bag of UTXOs will be generated. The number of UTXOs and the distribution of their values and script types will be modifiable.
- A random bag of payment requests will also be generated. Each payment request is just a output UTXO that a wallet will need to create a payment for.
- Then, each wallet will be initialized with the same bag of UTXOs.
- And each wallet will receive payment requests (through the `SpendFromWallet` function) in the same order until it can no longer satisfy a payment.
- The number of total payments completed along with the effective remaining wallet balance will be used to determine the winning wallet implementation.
