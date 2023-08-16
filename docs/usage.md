# ethdo commands

ethdo provides features to manage wallets and accounts, as well as interacting with Ethereum consensus nodes and remote signers.  Below are a list of all available commands.

Note that the below provides a list of commands rather than a guide for specific situations.  Other guides available are:

- [How to change withdrawal credentials for a validator](./changingwithdrawalcredentials.md)
- [How to convert from mnemonics to keys and accounts](./conversions.md)
- [How to achieve common tasks with ethdo](./howto.md)

### `wallet` commands

#### `accounts`

`ethdo wallet accounts` lists the accounts within a wallet.  

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

#### `batch`

`ethdo wallet batch` batches the accounts in a wallet into a single file to allow faster decryption. Options for batching a wallet include:

- `wallet`: the name of the wallet to batch
- `passphrase`: the passphrase for of the accounts in the wallet
- `batch-passphrase`: the passphrase for the batch.   Note that this can be the same as the passphrase for the accounts

```sh
$ ethdo wallet batch --wallet="Validators" ---passphrase="my account secret" --batch-passphrase="my batch secret"
```

#### `create`

`ethdo wallet create` creates a new wallet with the given parameters.  Options for creating a wallet include:

- `wallet`: the name of the wallet to create
- `type`: the type of wallet to create.  This can be either "nd" for a non-deterministic wallet, where private keys are generated randomly, or "hd" for a hierarchical deterministic wallet, where private keys are generated from a seed and path as per [EIP-2333](https://eips.ethereum.org/EIPS/eip-2333) (defaults to "nd")
- `wallet-passphrase`: the passphrase for of the wallet.  This is required for hierarchical deterministic wallets, to protect the seed
- `mnemonic`: for hierarchical deterministic wallets only, use a pre-defined 24-word [BIP-39 seed phrase](https://en.bitcoin.it/wiki/Seed_phrase) to create the wallet, along with an additional "seed extension" phrase if required.  **Warning** The same mnemonic can be used to create multiple wallets, in which case they will generate the same keys.

```sh
$ ethdo wallet create --wallet="Personal wallet" --type="hd" --wallet-passphrase="my wallet secret"
```

#### `delete`
`ethdo wallet delete` deletes a wallet.  Options for deleting a wallet include:

- `wallet`: the name of the wallet to delete

```sh
$ ethdo wallet delete --wallet="Old wallet"
```

**Warning** Deleting a wallet is permanent.  Only use this command if you really don't want the wallet, or you have securely backed the wallet up using `wallet export`.

#### `export`

`ethdo wallet export` exports the wallet and all of its accounts.  Options for exporting a wallet include:

- `wallet`: the name of the wallet to export (defaults to "primary")
- `passphrase`: the passphrase with which to encrypt the wallet backup

```sh
$ ethdo wallet export --wallet="Personal wallet" --passphrase="my export secret"
0x01c7a27ad40d45b4ae5be5f...
```

The encrypted wallet export is written to the console; it can be redirected to store it in a file.

```sh
$ ethdo wallet export --wallet="Personal wallet" --passphrase="my export secret" >export.dat
```

#### `import`

`ethdo wallet import` imports a wallet and all of its accounts exported by `ethdo wallet export`.  Options for importing a wallet include:

- `data`: the data exported by `ethdo wallet export`
- `passphrase`: the passphrase that was provided to `ethdo wallet export` to encrypt the data
- `verify`: confirm information about the wallet import without importing it

```sh
$ ethdo wallet import --data="0x01c7a27ad40d45b4ae5be5f..." --passphrase="my export secret"
```

The encrypted wallet export can be read from a file.  For example with Unix systems:

```sh
$ ethdo wallet import --data=`cat export.dat` --passphrase="my export secret"
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

#### `sharedexport`

`ethdo wallet sharedexport` exports the wallet and all of its accounts with shared keys.  Options for exporting a wallet include:

- `wallet`: the name of the wallet to export (defaults to "primary")
- `participants`: the total number of participants that each hold a share
- `threshold`: the number of participants necessary to provide their share to restore the wallet
- `file`: the name of the file that stores the backup

```sh
$ ethdo wallet sharedexport --wallet="Personal wallet" --participants=3 --threshold=2 --file=backup.dat
298a4efce34c7f46114b7c4ea4be3d3bef925dccb153dd00227b53c3be7dad668b326f2659b2375e708bb824b33f0e7364e0f21dd18e5f5f2d7d04de7c122a9189
10eabc645fd1633e2e874ffc486fcbe313b33c34fbcb30511a74517296e01f332c6e3c40757c39b4dc47f3a417d321c4c81c2115e53fca57797c975913b8bf5063
559c77b56d36cbd84f23669a376d389f8c6644933a1a4112512e4d063d0779489c6b6312e6c46def0ee33ce5b5aca0941a833f65e64b5d270c7224323f4e28b238
```

Each line of the output is a share and should be provided to one of the participants, along with the backup file.

#### `sharedimport`

`ethdo wallet sharedimport` imports a wallet and all of its accounts exported by `ethdo wallet sharedexport`.  Options for importing a wallet include:

- `file`: the name of the file that stores the backup
- `shares`: a number of shares, defined by _threshold_ during the export, separated by spaces

```sh
$ ethdo wallet sharedimport --file=backup.dat --shares="298a…9189 10ea…5063"
```

### `account` commands

Account commands focus on information about local accounts, generally those used by Geth and Parity but also those from hardware devices.

#### `create`

`ethdo account create` creates a new account with the given parameters.  Options for creating an account include:

- `account`: the name of the account to create (in format "wallet/account")
- `passphrase`: the passphrase for the account
- `path`: the HD path for the account (only for hierarchical deterministic accounts)

Note that for hierarchical deterministic wallets you will also need to supply `--wallet-passphrase` to unlock the wallet seed.

For distributed accounts you will also need to supply `--participants` and `--signing-threshold`.

```sh
$ ethdo account create --account="Personal wallet/Operations" --wallet-passphrase="my wallet secret" --passphrase="my account secret"
```

#### `derive`

`ethdo account derive` provides the ability to derive an account's keys without creating either the wallet or the account.  This allows users to quickly obtain or confirm keys without going through a relatively long process, and has the added security benefit of not writing any information to disk.  Options for deriving the account include:

- `mnemonic`: a pre-defined 24-word [BIP-39 seed phrase](https://en.bitcoin.it/wiki/Seed_phrase) to derive the account, along with an additional "seed extension" phrase if required supplied as the 25th word
- `path`: the HD path used to derive the account
- `show-private-key`: show the private of the derived account.  **Warning** displaying private keys, especially those derived from seeds held on hardware wallets, can expose your Ether to risk of being stolen.  Only use this option if you are sure you understand the risks involved
- `show-withdrawal-credentials`: show the withdrawal credentials of the derived account
- `generate-keystore`: generate a keystore for the account

```sh
$ ethdo account derive --mnemonic="abandon ... abandon art" --path="m/12381/3600/0/0"
Public key: 0x99b1f1d84d76185466d86c34bde1101316afddae76217aa86cd066979b19858c2c9d9e56eebc1e067ac54277a61790db
```

#### `import`

`ethdo account import` creates a new account by importing its private key.  Options for creating the account include:

- `account`: the name of the account to create (in format "wallet/account")
- `passphrase`: the passphrase for the account
- `key`: the private key to import

```sh
$ ethdo account import --account=Validators/123 --key=6dd12d588d1c05ba40e80880ac7e894aa20babdbf16da52eae26b3f267d68032 --passphrase="my account secret"
```

You can also import from an existing keystore such as those generated by the deposit CLI.  For this you need the keystore and the keystore passphrase.  For example:

```sh
$ ethdo account import --account=Validators/123 --keystore=/path/to/keystore.json --keystore-passphrase="the keystore secret" --passphrase="my account secret"
```

`--keystore` can either be the path to the keystore file, or the contents of the keystore file.

#### `info`

`ethdo account info` provides information about the given account.  Options include:

- `account`: the name of the account on which to obtain information (in format "wallet/account")

```sh
$ ethdo account info --account="Personal wallet/Operations"
Public key: 0x8e2f9e8cc29658ff37ecc30e95a0807579b224586c185d128cb7a7490784c1ad9b0ab93dbe604ab075b40079931e6670
```

#### `key`

`ethdo account key` provides the private key for an account.  Options include:

- `account`: the name of the account on which to obtain information (in format "wallet/account")
- `passphrase`: the passphrase for the account

```sh
$ ethdo account key --account=interop/00001 --passphrase=secret
0x51d0b65185db6989ab0b560d6deed19c7ead0e24b9b6372cbecb1f26bdfad000
```

#### `lock`

`ethdo account lock` manually locks an account on a remote signer.  Locked accounts cannot carry out signing requests.  Options include:

- `account`: the name of the account to lock (in format "wallet/account")

Note that this command only works with remote signers; it has no effect on local accounts.

```sh
$ ethdo account lock --account=Validators/123
```

#### `unlock`

`ethdo account unlock` manually unlocks an account on a remote signer.  Unlocked accounts cannot carry out signing requests.  Options include:

- `account`: the name of the account to unlock (in format "wallet/account")
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
- `account`: the account to sign the data (in format "wallet/account")
- `passphrase`: the passphrase for the account

```sh
$ ethdo signature sign --data="0x08140077a94642919041503caf5cc1c89c7744a2a08d43cec91df1795b23ecf2" --account="Personal wallet/Operations" --passphrase="my account secret"
0x87c83b31081744667406a11170c5585a11195621d0d3f796bd9006ac4cb5f61c10bf8c5b3014cd4f792b143a644cae100cb3155e8b00a961287bd9e7a5e18cb3b80930708bc9074d11ff47f1e8b9dd0b633e71bcea725fc3e550fdc259c3d130
```

#### `signature verify`

`ethdo signature verify` verifies signed data.  Options include:

- `data`: the data whose signature to verify, as a hex string
- `signature`: the signature to verify, as a hex string
- `account`: the account which signed the data (if available as an account, in format "wallet/account")
- `signer`: the public key of the account which signed the data (if not available as an account)

```sh
$ ethdo signature verify --data="0x08140077a94642919041503caf5cc1c89c7744a2a08d43cec91df1795b23ecf2" --signature="0x87c83b31081744667406a11170c5585a11195621d0d3f796bd9006ac4cb5f61c10bf8c5b3014cd4f792b143a644cae100cb3155e8b00a961287bd9e7a5e18cb3b80930708bc9074d11ff47f1e8b9dd0b633e71bcea725fc3e550fdc259c3d130" --account="Personal wallet/Operations"
$ ethdo signature verify --data="0x08140077a94642919041503caf5cc1c89c7744a2a08d43cec91df1795b23ecf2" --signature="0x87c83b31081744667406a11170c5585a11195621d0d3f796bd9006ac4cb5f61c10bf8c5b3014cd4f792b143a644cae100cb3155e8b00a961287bd9e7a5e18cb3b80930708bc9074d11ff47f1e8b9dd0b633e71bcea725fc3e550fdc259c3d130" --account="Personal wallet/Auctions"
Not verified
$ ethdo signature verify --data="0x08140077a94642919041503caf5cc1c89c7744a2a08d43cec91df1795b23ecf2" --signature="0x89abe2e544ef3eafe397db036103b1d066ba86497f36ed4ab0264162eadc89c7744a2a08d43cec91df128660e70ecbbe11031b4c2e53682d2b91e67b886429bf8fac9bad8c7b63c5f231cc8d66b1377e06e27138b1ddc64b27c6e593e07ebb4b" --signer="0x8e2f9e8cc29658ff37ecc30e95a0807579b224586c185d128cb7a7490784c1ad9b0ab93dbe604ab075b40079931e6670"
$ ethdo signature verify --data="0x08140077a94642919041503caf5cc1c89c7744a2a08d43cec91df1795b23ecf2" --signature="0x87c83b31081744667406a11170c5585a11195621d0d3f796bd9006ac4cb5f61c10bf8c5b3014cd4f792b143a644cae100cb3155e8b00a961287bd9e7a5e18cb3b80930708bc9074d11ff47f1e8b9dd0b633e71bcea725fc3e550fdc259c3d130" --signer="0xad1868210a0cff7aff22633c003c503d4c199c8dcca13bba5b3232fc784d39d3855936e94ce184c3ce27bf15d4347695" --verbose
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

Block commands focus on providing information about Ethereum consensus blocks.
#### `analyze`
`ethdo block analyze` obtains information about a block in the Ethereum consensus chain.  Options include:

- `blockid`: the ID (slot, root, 'head') of the block to obtain

```sh
$ ethdo block analyze --blockid=80
Value for block 80: 488.531
```

Additional information is supplied when using `--verbose`

```
$ ethdo block analyze --blockid=80 --verbose
Attestation 0: distance 1, 119/119/132 new/total/possible votes, score 0.844, value 100.406
Attestation 1: distance 1, 116/116/131 new/total/possible votes, score 0.844, value 97.875
Attestation 2: distance 1, 115/115/131 new/total/possible votes, score 0.844, value 97.031
Attestation 3: distance 1, 114/114/132 new/total/possible votes, score 0.844, value 96.188
Attestation 4: distance 1, 113/113/132 new/total/possible votes, score 0.844, value 95.344
Attestation 5: distance 1, 2/22/132 new/total/possible votes, score 0.844, value 1.688
Value for block 80: 488.531
```

#### `info`

`ethdo block info` obtains information about a block in the Ethereum consensus chain.  Options include:

- `blockid`: the ID (slot, root, 'head') of the block to obtain
- `block-time`: the time (unix timestamp in decimal or hex, or a time in format YYYY-MM-DDTHH:MM:SS) of the block to obtain

```sh
$ ethdo block info --blockid=80
Attestations: 1
Attester slashings: 0
Deposits: 0
Voluntary exits: 0
```

Additional information is supplied when using `--verbose`

```sh
$ ethdo block info --blockid=80 --verbose
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

Chain commands focus on providing information about Ethereum consensus chains.

#### `eth1votes`

`ethdo chain eth1votes` obtains information about the votes for the next Ethereum 1 block to be incorporated in to the chain for deposits.  Options include:

- `epoch` show the votes at the end of the given epoch
- `json` provide JSON output

```sh
$ ethdo chain eth1votes
Voting period: 6
Slots through period: 1000
Votes this period: 959
Leading vote is for block 0x0ae5716ac1906592dbfb243ccadf90191f706d6f8c925b4f2712d2e24687553a with 356 votes
```

Additional information is supplied when using `--verbose`

#### `info`

`ethdo chain info` obtains information about an Ethereum consensus chain.

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

#### `queues`

`ethdo chain queues` obtains the activation and exit queue lengths of an Ethereum chain from the node's point of view.  Options include:

- `epoch` show the queue length at a given epoch
- `json` provide JSON output

```sh
$ ethdo chain queues
Activation queue: 14798
```

#### `spec`

`ethdo chain spec` obtains the specification of an Ethereum consensus chain from the nod.

```sh
$ ethdo chain spec
ALTAIR_FORK_EPOCH: 74240
ALTAIR_FORK_VERSION: 0x01000000
BASE_REWARD_FACTOR: 64
...
```

#### `status`

`ethdo chain status` obtains the status of an Ethereum consensus chain from the node's point of view.  Options include:

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

#### `time`

`ethdo chain time` calculates the time period of Ethereum consensus epochs and slots.  Options include:

- `epoch` show epoch and slot times for the given epoch
- `slot` show epoch and slot times for the given slot
- `timestamp` show epoch and slot times for the given timestamp

```sh
$ ethdo chain time --epoch=1234
Epoch 1234
  Epoch start 2020-12-06 23:37:59
  Epoch end 2020-12-06 23:44:23
Slot 39488
  Slot start 2020-12-06 23:37:59
  Slot end 2020-12-06 23:38:11
```

### `deposit` comands

Deposit commands focus on information about deposit data information in a JSON file generated by the `ethdo validator depositdata` command.

#### `verify`

`ethdo deposit verify` verifies one or more deposit data information in a JSON file generated by the `ethdo validator depositdata` command.  Options include:

- `data`: either a path to the JSON file, the JSON itself, or a hex string representing a deposit transaction
- `withdrawalpubkey`: the public key of the withdrawal for the deposit.  If no value is supplied then withdrawal credentials for deposits will not be checked
- `validatorpubkey`: the public key of the validator for the deposit.  If no value is supplied then validator public keys will not be checked
- `depositvalue`: the value of the Ether being deposited.  If no value is supplied then deposit values will not be checked.

```sh
$ ethdo deposit verify --data=${HOME}/depositdata.json --withdrawalpubkey=0xad1868210a0cff7aff22633c003c503d4c199c8dcca13bba5b3232fc784d39d3855936e94ce184c3ce27bf15d4347695 --validatorpubkey=0xa951530887ae2494a8cc4f11cf186963b0051ac4f7942375585b9cf98324db1e532a67e521d0fcaab510edad1352394c --depositvalue=32Ether
```

### `epoch` comands

Epoch commands focus on information about a beacon chain epoch.

#### `summary`

`ethdo epoch summary` provides a summary of the given epoch.  Options include:

- `epoch`: the epoch for which to provide a summary; defaults to last complete epoch
- `json`: provide JSON output

```sh
$ ethdo epoch summary
Epoch 380:
  Proposals: 31/32 (96.88%)
  Attestations: 1530/1572 (97.33%)
  Sync committees: 13086/15872 (82.45%)
```

More detailed information can be obtained with the `--verbose` flag:

```sh
$ ethdo epoch summary --verbose
Epoch 380:
  Proposals: 31/32 (96.88%)
    Slot 12188 (28/32) validator 1518 not proposed or not included
  Attestations: 1530/1572 (97.33%)
    Slot 12160 committee 0 validator 292 failed to participate
    Slot 12162 committee 0 validator 204 failed to participate
    Slot 12163 committee 0 validator 297 failed to participate
    Slot 12164 committee 0 validator 209 failed to participate
    ...
```

### `exit` comands

Exit commands focus on information about validator exits generated by the `ethdo validator exit` command.

#### `verify`

`ethdo exit verify` verifies the validator exit information in a JSON file generated by the `ethdo validator exit` command.  Options include:

- `signed-operation`: either a path to the JSON file or the JSON itself

```sh
$ ethdo exit verify --signed-operation=${HOME}/exit.json
```

### `node` commands

Node commands focus on information from an Ethereum consensus node.

#### `events`

`ethdo node events` displays events emitted by an Ethereum consensus node.

```sh
$ ethdo node events --topics=head,chain_reorg
{"topic":"head","data":{"block":"0x3b98ccd8dbd39e0763a5e91ef754f4559f5f9b6b8014ff45c8abf2eaf236324a","current_duty_dependent_root":"0xdedb063fcc2f404c701fe05a5bfbc95881d23292239adc1c4fc49d409beea7be","epoch_transition":false,"previous_duty_dependent_root":"0x626037a43b204b88d1911510d2717669ee2839b3227f5642832c36e5e6fc5e2f","slot":"231380","state":"0x009630f55970fa73b61cc97b0ed83b2faaf9103c8db722705e1b4376f240e415"}}
{"topic":"head","data":{"block":"0x41e4c09da2ddbca777fbdae5b0cdd2823df630fb74e1d3f6837586046737b414","current_duty_dependent_root":"0xdedb063fcc2f404c701fe05a5bfbc95881d23292239adc1c4fc49d409beea7be","epoch_transition":false,"previous_duty_dependent_root":"0x626037a43b204b88d1911510d2717669ee2839b3227f5642832c36e5e6fc5e2f","slot":"231381","state":"0xd2ce52483df1add63df441c15d4aef291be8a41d6d1d3ce7e9d6d619f7c911de"}}
...
```

#### `info`

`ethdo node info` obtains the information about an Ethereum consensus node.

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

### `slot` commands

Slot commands focus on information about Ethereum consensus slots.

#### `slottime`

`ethdo slot time` provides information about the time of a slot.  options include:

- `slot` the slot for which to provide the time

```sh
$ ethdo slot time --slot=5
2020-12-01 12:01:23 +0000 GMT
```

### `synccommittee` commands

Sync committee commands focus on information about sync committees.

#### `inclusion`

`ethdo synccommittee inclusion` provides information about the inclusion, or not, of a validator's sync committee messages.  Options include:

- `validator`: a [validator specifier](https://github.com/wealdtech/ethdo#validator-specifier)
- `epoch` the specific epoch for which to print sync committee contributions.  Defaults to the last complete epoch

```sh
$ ethdo synccommittee inclusion --index=274946 --epoch=91592
Epoch: 91593
Expected: 32
Included: 30
Missed: 1
No block: 1
Per-slot result: ✓✓✓✓✓✓✓✓ ✓✓✕✓✓✓✓✓ ✓✓✓✓-✓✓✓ ✓✓✓✓✓✓✓✓
```

#### `members`

`ethdo synccommittee members` provides information about the members of a sync committee.  Options include:

- `epoch` the specific epoch for which to provide sync committee members.
- `period` the period for which to provide sync committee members.  Can be 'current' or 'next'; dfeaults to 'current'

```sh
$ ethdo synccommittee members
138334,116317,231736,65706,60046,148162,274946,34724,18051,122841,269578,121110,89733,154887,202118,243459,267543,82793,59504,238929,55360,272874,93917,83116,264342,244312,264907,79193,15443,27997,127175,140965,64416,66399,173906,268885,67779,48139,215005,191435,107954,225228,148630,169357,61091,223319,40668,184307,95903,81179,237461,41723,119710,243333,248243,42757,228686,252749,17546,231625,132030,15934,108465,104302,93026,191946,63738,80996,90679,227542,75463,64581,242030,5429,61623,157314,145363,224733,232492,45357,80674,198583,221422,48665,154803,128608,172512,261074,102835,129935,255726,40846,218932,139874,194575,17346,171565,76413,237859,103170,95661,83018,73902,246680,35795,257792,23836,136624,45745,190990,124229,37281,23818,233435,253903,37502,8669,31151,267179,27954,181019,145719,112270,1899,184844,175014,121769,41717,218760,44813,255860,64865,31985,231664,134296,88114,185542,27557,1698,62470,79182,184325,80380,8865,218456,178979,243886,9466,221389,131476,160857,62916,195389,160182,99293,100263,242371,144594,227527,275978,65714,74350,60121,46642,219334,157142,99379,203508,84367,251808,276456,92563,199831,215312,193875,129690,104234,44290,227725,194780,163061,162328,176517,278620,137355,212826,131615,125734,151873,18977,147927,272759,160537,210675,180411,24203,37266,247527,128678,270287,90352,23043,169645,5304,183412,237387,79751,37635,275139,95857,185990,235565,49425,255836,254314,77582,104172,168556,143653,64173,64504,130363,216602,218107,181130,191845,56454,2040,270365,161952,222409,45097,51611,219190,154903,162311,257460,106337,110775,42928,275709,202352,54724,272295,274470,35220,19694,10347,169585,104938,35121,212982,190582,77999,110201,141519,239881,81263,84314,148883,254649,256309,270013,254179,134009,149660,177127,201926,30533,164789,154343,57437,28958,135169,186415,218514,171355,165247,213526,100044,184264,93278,269329,159634,4092,224671,217236,123946,80703,85444,247742,17959,146473,128231,167559,133899,181532,33378,79060,119785,249443,180469,43692,169679,154421,114047,87877,28337,59072,19807,204598,220293,99461,55272,227923,4503,12580,27044,68955,157373,61321,265034,106833,31534,69137,264783,129588,70433,88338,113528,226211,123003,118982,131549,60350,78896,165715,119736,52639,93274,164295,278837,186453,69910,36768,249533,106205,184057,253232,88155,121377,242589,148236,250065,191526,277249,157463,226527,93000,64784,176880,176380,144301,52061,169803,134291,96648,211716,223000,157911,256737,100938,50434,41075,114894,259888,116872,218201,83617,76348,256832,17113,50270,96468,128448,36987,127511,42397,10154,49234,193346,126352,57719,17029,213127,157942,187829,2353,62462,73637,29053,120324,108515,254684,35982,188131,217092,256206,85802,105907,21204,147562,188961,154541,131147,16000,225112,58362,170375,42239,188309,60280,125472,220119,268946,65736,274053,223569,60454,239552,4401,139357,279634,162711,112016,90295,170641,239770,212067,213770,78311,49057,256295,28666,167207,166783,213148,30689,72118,55912,197733,205116,106169,40570,225057,122079,126423,217781,212897,147499,201774,10616,157826,155954,258431,212151,255318,97138,151907,181491,40236,272993,104430,178068,56089,10067,185066,93669,124108,12785,230215,67995,196282,248285,215370,167715,186183,238147,164161,15068,127990,166146,244578,195912,199812,248435,135597,143024,225304,27045,238140,87008,272550,165234,218128,160038,17697,25332,23446,265921,201045,241106
```

### `validator` commands

Validator commands focus on interaction with Ethereum consensus validators.

#### `credentials get`

`ethdo validator credentials get` provides information about the withdrawal credentials for the provided validator.  Options include:

- `validator`: a [validator specifier](https://github.com/wealdtech/ethdo#validator-specifier)

```sh
$ ethdo validator credentials get --validator=Validators/1
```

#### `credentials set`

`ethdo validator credentials set` updates withdrawal credentials from BLS "type 0" credentials to execution "type 1" credentials.  Full information about using this command can be found in the [specific documentation](./changingwithdrawalcredentials.md).

```sh
$ ethdo validator credentials set --validator=Validators/1 --withdrawal-address=0x8f…9F --private-key=0x3b…9c
```

#### `depositdata`

`ethdo validator depositdata` generates the data required to deposit one or more Ethereum consensus validators.  Options include:

- `withdrawalaccount` specify the account to be used for the withdrawal credentials (if withdrawalpubkey is not supplied)
- `withdrawaladdress` specify the Ethereum execution address to be used for the withdrawal credentials (if withdrawalpubkey is not supplied)
- `withdrawalpubkey` specify the public key to be used for the withdrawal credentials (if withdrawalaccount is not supplied)
- `validatoraccount` specify the account to be used for the validator
- `depositvalue` specify the amount of the deposit
- `forkversion` specify the fork version for the deposit signature; this defaults to mainnet.  Note that supplying an incorrect value could result in the loss of your deposit, so only supply this value if you are sure you know what you are doing.  You can find the value for other chains by fetching the value supplied in "Genesis fork version" of the `ethdo chain info` command
- `raw` generate raw hex output that can be supplied as the data to an Ethereum 1 deposit transaction

#### `exit`

`ethdo validator exit` sends a transaction to the chain to tell an active validator to exit the validation queue.  Full information about using this command can be found in the [specific documentation](./exitingvalidators.md).

```sh
$ ethdo validator exit --validator=Validators/1 --passphrase="my validator secret"
```

To send a transaction when the account is not accessible to ethdo accout you can use the validator's private key instead:

```sh
$ ethdo validator exit --private-key=0x01e748d098d3bcb477d636f19d510399ae18205fadf9814ee67052f88c1f88c0
```

#### `info`

`ethdo validator info` provides information for a given validator.  Options include:

- `validator`: the validator for which to obtain information, as a [validator specifier](https://github.com/wealdtech/ethdo#validator-specifier)

```sh
$ ethdo validator info --validator=Validators/1
Status:            Active
Balance:           3.203823585 Ether
Effective balance: 3.1 Ether
```

Additional information is supplied when using `--verbose`

```sh
$ ethdo validator info --validator=0xb3bb6b7a8d809e59544472853d219499765bf01d14de1e0549bd6fc2a86627ac9033264c84cd503b6339e3334726562f --verbose
Epoch of data:          3398
Index:                  26913
Public key:             0xb3bb6b7a8d809e59544472853d219499765bf01d14de1e0549bd6fc2a86627ac9033264c84cd503b6339e3334726562f
Status:                 Active
Balance:                3.204026813 Ether
Effective balance:      3.1 Ether
Withdrawal credentials: 0x0033ef3cb10b36d0771ffe8a02bc5bfc7e64ea2f398ce77e25bb78989edbee36
```

#### `keycheck`

`ethdo validator keycheck` checks if a given key matches a validator's withdrawal credentials.  Options include:

- `withdrawal-credentials` the withdrawal credentials against which to match
- `private-key` the private key used to generat matching withdrawal credentials
- `mnemonic` the mnemonic used to generate matching withdrawal credentials

```sh
$ ethdo validator keycheck --withdrawal-credentials=0x007e28dcf9029e8d92ca4b5d01c66c934e7f3110606f34ae3052cbf67bd3fc02 --mnemonic='abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art'
Withdrawal credentials confirmed at path m/12381/3600/10/0
```

#### `expectation`

`ethdo validator expectation` calculates the times between expected actions.  Options include:

- `validators` the number of active validators to use as the basis for calculations

```sh
$ ethdo validator expectation
Expected time between block proposals: 4 weeks 6 days
Expected time between sync committees: 1 year 27 weeks
```

### `attester` commands

Attester commands focus on Ethereum consensus validators' actions as attesters.

#### `duties`

`ethdo attester duties` provides information on the duties that a given validator has in a given epoch.  Options include:

- `epoch` the epoch in which to obtain the duties (defaults to current epoch)
- `validator`: the validator for which to fetch the duties, as a [validator specifier](https://github.com/wealdtech/ethdo#validator-specifier)

```sh
$ ethdo attester duties --validator=Validators/0 --epoch=5
Validator attesting in slot 186 committee 3
```

#### `inclusion`

`ethdo attester inclusion` finds the block with wihch an attestation is included on the chain.  Options include:

- `epoch` the epoch in which to obtain the inclusion information (defaults to previous epoch)
- `validator`: the validator for which to fetch the duties, as a [validator specifier](https://github.com/wealdtech/ethdo#validator-specifier)

```sh
$ ethdo attester inclusion --validator=Validators/1 --epoch=6484
Attestation included in block 207492 (inclusion delay 1)
```

#### `withdrawal`
`ethdo validator withdrawal` provides information about the next withdrawal for the given validator.  Options include:

- `validator`: the validator for which to fetch the withdrawal, as a [validator specifier](https://github.com/wealdtech/ethdo#validator-specifier)
- `json`: provide JSON output

```sh
$ ethdo validator withdrawal --validator=12345
Withdrawal expected at 2023-04-17T15:08:35 in block 6243041
```

#### `yield`

`ethdo validator yield` calculates the expected yield given the number of validators.  Options include:

- `validators` use a specified number of validators rather than the current number of active validators
- `epoch` the epoch for which to calculate yield; defaults to the current epoch
- `json` obtain detailed information in JSON format

```sh
$ ethdo validator yield
Yield: 4.64%
```

#### `summary`
`ethdo validator summary` provides a summary of the given epoch for the given validators.  Options include:

- `epoch`: the epoch for which to provide a summary; defaults to last complete epoch
- `validators`: the list of validators for which to provide a summary, as [validator specifiers](https://github.com/wealdtech/ethdo#validator-specifier)
- `json`: provide JSON output

### `proposer` commands

Proposer commands focus on Ethereum consensus validators' actions as proposers.

#### `duties`

`ethdo proposer duties` provides information on the proposal duties for a given epoch.  Options include:

- `epoch` the epoch in which to obtain the duties (defaults to current epoch)
- `slot` the slot in which to obtain the duties (overrides epoch if present)
- `json` obtain detailed information in JSON format

```sh
$ ethdo proposer duties --epoch=5
Epoch 5:
  Slot 160: validator 8221
  Slot 161: validator 11193
  Slot 162: validator 4116
  Slot 163: validator 631
  ...
```

## Maintainers

Jim McDonald: [@mcdee](https://github.com/mcdee).

## Contribute

Contributions welcome. Please check out [the issues](https://github.com/wealdtech/ethdo/issues).

## License

[Apache-2.0](LICENSE) © 2019, 2020 Weald Technology Trading Ltd

