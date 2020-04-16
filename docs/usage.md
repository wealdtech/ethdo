# ethdo commands

ethdo provides features to manage wallets and accounts, as well as interacting with Ethereum 2 nodes and remote signers.  Below are a list of all available commands.

### `wallet` commands

#### `accounts`

`ethdo wallet accouts` lists the accounts within a wallet.  

```sh
$ ethdo wallet accounts --wallet="Personal wallet"
Auctions
Operations
Spending
```

With the `--verbose` flag this will provide the public key of the accounts.

```sh
$ ethdo wallet accounts --wallet="Personal wallet" --verbose
Auctions: 0x812f340269c315c1d882ae7c13cdaddf862dbdbd482b1836798b2070160dd1e194088cc6f39347782028d1e56bd18674
Operations: 0x8e2f9e8cc29658ff37ecc30e95a0807579b224586c185d128cb7a7490784c1ad9b0ab93dbe604ab075b40079931e6670
Spending: 0x85dfc6dcee4c9da36f6473ec02fda283d6c920c641fc8e3a76113c5c227d4aeeb100efcfec977b12d20d571907d05650
```
#### `create`

`ethdo wallet create` creates a new wallet with the given parameters.  Options for creating a wallet include:
  - `wallet`: the name of the wallet to create (defaults to "primary")
  - `type`: the type of wallet to create.  This can be either "nd" for a non-deterministic wallet, where private keys are generated randomly, or "hd" for a hierarchical deterministic wallet, where private keys are generated from a seed and path as per [ERC-2333](https://github.com/CarlBeek/EIPs/blob/bls_path/EIPS/eip-2334.md) (defaults to "nd")
  - `walletpassphrase`: the passphrase for of the wallet.  This is required for hierarchical deterministic wallets, to protect the seed

```sh
$ ethdo wallet create --wallet="Personal wallet" --type="hd" --walletpassphrase="my wallet secret"
```

#### `export`

`ethdo wallet export` exports the wallet and all of its accounts.  Options for exporting a wallet include:
  - `wallet`: the name of the wallet to export (defaults to "primary")
  - `exportpassphrase`: the passphrase with which to encrypt the wallet backup

```sh
$ ethdo wallet export --wallet="Personal wallet" --exportpassphrase="my export secret"
0x01c7a27ad40d45b4ae5be5f...
```

The encrypted wallet export is written to the console; it can be redirected to store it in a file.

```sh
$ ethdo wallet export --wallet="Personal wallet" --exportpassphrase="my export secret" >export.dat
```

#### `import`

`ethdo wallet import` imports a wallet and all of its accounts exported by `ethdo wallet export`.  Options for importing a wallet include:
  - `importdata`: the data exported by `ethdo wallet export`
  - `importpassphrase`: the passphrase that was provided to `ethdo wallet export` to encrypt the data

```sh
$ ethdo wallet import --importdata="0x01c7a27ad40d45b4ae5be5f..." --importpassphrase="my export secret"
```

The encrypted wallet export can be read from a file.  For example with Unix systems:

```sh
$ ethdo wallet import --importdata=`cat export.dat` --importpassphrase="my export secret"
```

#### `info`

`ethdo wallet info` provides information about a given wallet.  Options include:
  - `wallet`: the name of the wallet

```sh
$ ethdo wallet info --wallet="Personal wallet"
Type: hierarchical deterministic
Accounts: 3
```

#### `list`

`ethdo wallet list` lists all wallets in the store.

```sh
$ ethdo wallet list
Personal wallet
```

**N.B.** encrypted wallets will not show up in this list unless the correct passphrase for the store is supplied.

#### `seed`

`ethdo wallet seed` provides the seed for hierarchical deterministic wallets.  Options include:
  - `wallet`: the name of the wallet
  - `walletpassphrase`: the passphrase for the wallet

```sh
$ ethdo wallet seed --wallet="Personal wallet" --walletpassphrase="my wallet secret"
decorate false mail domain gain later motion chair tank muffin smoke involve witness bean shell urge team solve share truly shadow decorate jeans hen
```

### `account` commands

Account commands focus on information about local accounts, generally those used by Geth and Parity but also those from hardware devices.

#### `create`

`ethdo account create` creates a new account with the given parameters.  Options for creating an account include:
  - `account`: the name of the account to create
  - `passphrase`: the passphrase for the account

Note that for hierarchical deterministic wallets you will also need to supply `--walletpassphrase` to unlock the wallet seed.

```sh
$ ethdo account create --account="Personal wallet/Operations" --walletpassphrase="my wallet secret" --passphrase="my account secret"
```
#### `import`

`ethdo account import` creates a new account by importing its private key.  Options for creating the account include:
  - `account`: the name of the account to create
  - `passphrase`: the passphrase for the account
  - `key`: the private key to import

```sh
$ ethdo account import --account=Validators/123 --key=6dd12d588d1c05ba40e80880ac7e894aa20babdbf16da52eae26b3f267d68032 --passphrase="my account secret"
```

#### `info`

`ethdo account info` provides information about the given account.  Options include:
  - `account`: the name of the account on which to obtain information

```sh
$ ethdo account info --account="Personal wallet/Operations"
Public key: 0x8e2f9e8cc29658ff37ecc30e95a0807579b224586c185d128cb7a7490784c1ad9b0ab93dbe604ab075b40079931e6670
```

#### `key`

`ethdo account key` provides the private key for an account.  Options include:
  - `account`: the name of the account on which to obtain information
  - `passphrase`: the passphrase for the account

```sh
$ ethdo account key --account=interop/00001 --passphrase=secret
0x51d0b65185db6989ab0b560d6deed19c7ead0e24b9b6372cbecb1f26bdfad000
```

#### `lock`

`ethdo account lock` manually locks an account on a remote signer.  Locked accounts cannot carry out signing requests.  Options include:
  - `account`: the name of the account to lock

Note that this command only works with remote signers; it has no effect on local accounts.

```sh
$ ethdo account lock --account=Validators/123
```

#### `unlock`

`ethdo account unlock` manually unlocks an account on a remote signer.  Unlocked accounts cannot carry out signing requests.  Options include:
  - `account`: the name of the account to unlock
  - `passphrase`: the passphrase for the account

Note that this command only works with remote signers; it has no effect on local accounts.

```sh
$ ethdo account unlock --account=Validators/123 --passphrase="my secret passphrase"
```

### `signature` commands

Signature commands focus on generation and verification of data signatures.

#### `signature sign`

`ethdo signature sign` signs provided data.  Options include:
  - `data`: the data to sign, as a hex string
  - `domain`: the domain in which to sign the data.  This is a 32-byte hex string
  - `account`: the account to sign the data
  - `passphrase`: the passphrase for the account

```sh
$ ethdo signature sign --data="0x08140077a94642919041503caf5cc1795b23ecf2" --account="Personal wallet/Operations" --passphrase="my account secret"
0x89abe2e544ef3eafe397db036103b1d066ba86497f36ed4ab0264162eadc89c7744a2a08d43cec91df128660e70ecbbe11031b4c2e53682d2b91e67b886429bf8fac9bad8c7b63c5f231cc8d66b1377e06e27138b1ddc64b27c6e593e07ebb4b
```

#### `signature verify`

`ethdo signature verify` verifies signed data.  Options include:
  - `data`: the data whose signature to verify, as a hex string
  - `signature`: the signature to verify, as a hex string
  - `account`: the account which signed the data (if available as an account)
  - `signer`: the public key of the account which signed the data (if not available as an account)

```sh
$ ethdo signature verify --data="0x08140077a94642919041503caf5cc1795b23ecf2" --signature="0x89abe2e544ef3eafe397db036103b1d066ba86497f36ed4ab0264162eadc89c7744a2a08d43cec91df128660e70ecbbe11031b4c2e53682d2b91e67b886429bf8fac9bad8c7b63c5f231cc8d66b1377e06e27138b1ddc64b27c6e593e07ebb4b" --account="Personal wallet/Operations"
Verified
$ ethdo signature verify --data="0x08140077a94642919041503caf5cc1795b23ecf2" --signature="0x89abe2e544ef3eafe397db036103b1d066ba86497f36ed4ab0264162eadc89c7744a2a08d43cec91df128660e70ecbbe11031b4c2e53682d2b91e67b886429bf8fac9bad8c7b63c5f231cc8d66b1377e06e27138b1ddc64b27c6e593e07ebb4b" --account="Personal wallet/Auctions"
Not verified
$ ethdo signature verify --data="0x08140077a94642919041503caf5cc1795b23ecf2" --signature="0x89abe2e544ef3eafe397db036103b1d066ba86497f36ed4ab0264162eadc89c7744a2a08d43cec91df128660e70ecbbe11031b4c2e53682d2b91e67b886429bf8fac9bad8c7b63c5f231cc8d66b1377e06e27138b1ddc64b27c6e593e07ebb4b" --signer="0x8e2f9e8cc29658ff37ecc30e95a0807579b224586c185d128cb7a7490784c1ad9b0ab93dbe604ab075b40079931e6670"
Verified
```

The same rules apply to `ethereal signature verify` as those in `ethereal signature sign` above.

### `version`

`ethdo version` provides the current version of ethdo.  For example:

```sh
$ ethdo version
1.4.0
```

### `block` commands

Block commands focus on providing information about Ethereum 2 blocks.
#### `info`

`ethdo block info` obtains information about a block in Ethereum 2.  Options include:
  - `slot`: the slot at which to attempt to fetch the block

```sh
$ ethdo block info --slot=80 
Attestations: 1
Attester slashings: 0
Deposits: 0
Voluntary exits: 0
```

Additional information is supplied when using `--verbose`

```sh
$ ethdo block info --slot=80 --verbose
Parent root: 0x9a08aab7d5bbc816a9d2c20c79895519da2045e99ac6782ab3d05323a395fe51
State root: 0xc6a2626ba5cb37f984bdc4da4dc93a5012be5b69fdcebc50be70a1181a290265
Ethereum 1 deposit count: 512
Ethereum 1 deposit root: 0x05b88acdde2092e1ecf35714dca0ccf82fb7e73180643f51d3139553136d125f
Ethereum 1 block hash: 0x2b8d87e016376d83b2c04c1e626172a3f8bef3b4a37d7f2f3f76d0c62acdf573
Attestations: 1
	0:
		Committee index: 0
		Attesters: 17
		Aggregation bits: ✓✓✓✓✓✓✓✓ ✓✓✓✓✓✓✓✓ ✕✕✕✕✕✕✕✓
		Slot: 79
		Beacon block root: 0x9a08aab7d5bbc816a9d2c20c79895519da2045e99ac6782ab3d05323a395fe51
		Source epoch: 0
		Source root: 0x0000000000000000000000000000000000000000000000000000000000000000
		Target epoch: 2
		Target root: 0xb93273c516fc817e64fab53ff4093f295e5da463582e85e1ca60800e9464faf2
Attester slashings: 0
Deposits: 0
Voluntary exits: 0
```

### `chain` commands

Chain commands focus on providing information about Ethereum 2 chains.

#### `info`

`ethdo chain info` obtains information about an Ethereum 2 chain.

```sh
$ ethdo chain info
Genesis time:		Thu Apr 16 08:02:43 BST 2020
```

Additional information is supplied when using `--verbose`

```sh
$ ethdo chain info --verbose
Genesis time:		Thu Apr 16 08:02:43 BST 2020
Genesis fork version:	00000000
Seconds per slot:	12
Slots per epoch:	32
```

#### `status`

`ethdo chain status` obtains the status of an Ethereum 2 chain from the node's point of view.  Options include:
  - `slot` show output in terms of slots rather than epochs

```sh
$ ethdo chain status
Current epoch: 5
Justified epoch: 4
Finalized epoch: 3
```

Additional information is supplied when using `--verbose`

```sh
$ ethdo chain status --verbose
Current epoch: 5
Justified epoch: 4
Justified epoch distance 1
Finalized epoch: 3
Finalized epoch distance: 2
Prior justified epoch: 3
Prior justified epoch distance: 4
```

### `node` commands

Node commands focus on information from an Ethereum 2 node.

#### `info`

`ethdo node info` obtains the information about an Ethereum 2 node.

```sh
$ ethdo node info
Syncing: false
Current slot: 178
Current epoch: 5
```

Additional information is supplied when using `--verbose`

```sh
$ ethdo node info --verbose
Version: Prysm/Git commit: b0aa6e22455e4d9cb8720a259771fbbbd22dc3ec. Built at: 2020-04-16T08:02:43+01:00
Syncing: false
Current slot: 178
Current epoch: 5
Genesis timestamp: 1587020563
```
### `validator` commands

Validator commands focus on interaction with Ethereum 2 validators.

#### `depositdata`

`ethdo validator depositdata` generates the data required to deposit one or more Ethereum 2 validators.

#### `exit`

`ethdo validator exit` sends a transaction to the chain to tell an active validator to exit the validation queue.

```sh
$ ethdo validator exit --account=Validators/1 --passphrase="my validator secret"
```

#### `info`

`ethdo validator info` provides information for a given validator.

```sh
$ ethdo validator info --account=Validators/1
Status:            Active
Balance:           3.203823585 Ether
Effective balance: 3.1 Ether
```

Additional information is supplied when using `--verbose`

```sh
$ ethdo validator info --account=Validators/1 --verbose
Epoch of data:          3398
Index:                  26913
Public key:             0xb3bb6b7a8d809e59544472853d219499765bf01d14de1e0549bd6fc2a86627ac9033264c84cd503b6339e3334726562f
Status:                 Active
Balance:                3.204026813 Ether
Effective balance:      3.1 Ether
Withdrawal credentials: 0x0033ef3cb10b36d0771ffe8a02bc5bfc7e64ea2f398ce77e25bb78989edbee36
```

If the validator is not an account it can be queried directly with `--pubkey`.

```sh
$ ethdo validator info --pubkey=0x842dd66cfeaeff4397fc7c94f7350d2131ca0c4ad14ff727963be9a1edb4526604970df6010c3da6474a9820fa81642b
Status:            Active
Balance:           3.201850307 Ether
Effective balance: 3.1 Ether
```

## Maintainers

Jim McDonald: [@mcdee](https://github.com/mcdee).

## Contribute

Contributions welcome. Please check out [the issues](https://github.com/wealdtech/ethdo/issues).

## License

[Apache-2.0](LICENSE) © 2019, 2020 Weald Technology Trading Ltd

