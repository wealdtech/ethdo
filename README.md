[![Tag](https://img.shields.io/github/tag/wealdtech/ethdo.svg)](https://github.com/wealdtech/ethdo/releases/)
[![License](https://img.shields.io/github/license/wealdtech/ethdo.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/wealdtech/ethdo?status.svg)](https://godoc.org/github.com/wealdtech/ethdo)
[![Travis CI](https://img.shields.io/travis/wealdtech/ethdo.svg)](https://travis-ci.org/wealdtech/ethdo)

A command-line tool for managing common tasks in Ethereum 2.

** Please note that this library uses standards that are not yet final, and as such may result in changes that alter public and private keys.  Do not use this library for production use just yet **

## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contribute](#contribute)
- [License](#license)

## Install


`ethdo` is a standard Go program which can be installed with:

```sh
go get github.com/wealdtech/ethdo
```

## Usage

ethdo contains a large number of features that are useful for day-to-day interactions with the Ethereum 2 blockchain.

### Wallets and accounts

ethdo uses the [go-eth2-wallet](https://github.com/wealdtech/go-eth2-wallet) system to provide unified access to different wallet types.

All ethdo comands take the following parameters:

  - `store`: the name of the storage system for wallets.  This can be one of "filesystem" or "s3", and defaults to "filesystem"
  - `storepassphrase`: the passphrase for the store.  If this is empty the store is unencrypted
  - `walletpassphrase`: the passphrase for the wallet.  This is required for some wallet-centric operations such as creating new accounts
  - `accountpassphrase`: the passphrase for the account.  This is required for some account-centric operations such as signing data

Accounts are specified in the standard "<wallet>/<account>" format, for example the account "savings" in the wallet "primary" would be referenced as "primary/savings".

### Configuration file and environment

ethdo supports a configuration file; by default in the user's home directory but changeable with the `--config` argument on the command line.  The configuration file provides values that override the defaults but themselves can be overridden with command-line arguments.

The default file name is `.ethdo.json` or `.ethdo.yml` depending on the encoding used (JSON or YAML, respectively).  An example `.ethdo.json` file is shown below:

```json
{
  "store": "s3",
  "storepassphrase": "s3 secret passphrse",
  "account": "Personal wallet/Operations",
  "verbose": true
}
```

ethdo also supports environment variables.  Environment variables are prefixed with "ETHDO_" and are upper-cased.  So for example to provide your account passphrase in an environment variable on a Unix system you could use:

```sh
export ETHDO_PASSPHRASE="my account passphrase"
```

### Output and exit status

If set, the `--quiet` argument will suppress all output.

If set, the `--verbose` argument will output additional information related to the command.  Details of the additional information is command-specific and explained in the command help below.

If set, the `--debug` argument will output additional information about the operation of ethdo as it carries out its work.

Commands will have an exit status of 0 on success and 1 on failure.  The specific definition of success is specified in the help for each command.

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

#### `info`

`ethdo account info` provides information about the given account.  Options include:
  - `account`: the name of the account on which to obtain information

```sh
$ ethdo account info --account="Personal wallet/Operations"
Public key: 0x8e2f9e8cc29658ff37ecc30e95a0807579b224586c185d128cb7a7490784c1ad9b0ab93dbe604ab075b40079931e6670
```

### `signature` commands

Signature commands focus on generation and verification of data signatures.

### `signature sign`

`ethdo signature sign` signs provided data.  Options include:
  - `data`: the data to sign, as a hex string
  - `domain`: the domain in which to sign the data.  This is an 8-byte hex string (default 0x0000000000000000)
  - `account`: the account to sign the data
  - `passphrase`: the passphrase for the account

```sh
$ ethdo signature sign --data="0x08140077a94642919041503caf5cc1795b23ecf2" --account="Personal wallet/Operations" --passphrase="my account secret"
0x89abe2e544ef3eafe397db036103b1d066ba86497f36ed4ab0264162eadc89c7744a2a08d43cec91df128660e70ecbbe11031b4c2e53682d2b91e67b886429bf8fac9bad8c7b63c5f231cc8d66b1377e06e27138b1ddc64b27c6e593e07ebb4b
```

### `signature verify`

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
1.0.0
```

## Maintainers

Jim McDonald: [@mcdee](https://github.com/mcdee).

## Contribute

Contributions welcome. Please check out [the issues](https://github.com/wealdtech/ethdo/issues).

## License

[Apache-2.0](LICENSE) © 2019 Weald Technology Trading Ltd

