[![Tag](https://img.shields.io/github/tag/wealdtech/ethdo.svg)](https://github.com/wealdtech/ethdo/releases/)
[![License](https://img.shields.io/github/license/wealdtech/ethdo.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/wealdtech/ethdo?status.svg)](https://godoc.org/github.com/wealdtech/ethdo)
[![Travis CI](https://img.shields.io/travis/wealdtech/ethdo.svg)](https://travis-ci.org/wealdtech/ethdo)

A command-line tool for managing common tasks in Ethereum 2.

## Table of Contents

- [Install](#install)
  - [Binaries](#binaries)
  - [Docker](#docker)
  - [Source](#source)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contribute](#contribute)
- [License](#license)

## Install

### Binaries

Binaries for the latest version of `ethdo` can be obtained from [the releases page](https://github.com/wealdtech/ethdo/releases).

### Docker

You can obtain the latest version of `ethdo` using docker with:

```
docker pull wealdtech/ethdo
```

### Source
`ethdo` is a standard Go program which can be installed with:

```sh
GO111MODULE=on go get github.com/wealdtech/ethdo
```

Note that `ethdo` requires at least version 1.13 of go to operate.  The version of go can be found with `go version`.

If this does not work please see the [troubleshooting](https://github.com/wealdtech/ethdo/blob/master/docs/troubleshooting.md) page.

The docker image can be build locally with:

```sh
docker build -t ethdo .
```

You can run `ethdo` using docker after that. Example:

```sh
docker run -it ethdo --help
```

Note that that many `ethdo` commands connect to the beacon node to obtain information.  If the beacon node is running directly on the server this requires the `--network=host` command, for example:

```sh
docker run --network=host ethdo chain status
```

Alternatively, if the beacon node is running in a separate docker container a shared network can be created with `docker network create eth2` and accessed by adding `--network=eth2` added to both the beacon node and `ethdo` containers.

## Usage

ethdo contains a large number of features that are useful for day-to-day interactions with the Ethereum 2 blockchain.

### Wallets and accounts

ethdo uses the [go-eth2-wallet](https://github.com/wealdtech/go-eth2-wallet) system to provide unified access to different wallet types.  When on the filesystem the locations of the created wallets and accounts are:

    - for Linux: $HOME/.config/ethereum2/wallets
    - for OSX: $HOME/Library/Application Support/ethereum2/wallets
    - for Windows: %APPDATA%\ethereum2\wallets

If using the filesystem store, the additional parameter `basedir` can be supplied to change this location.

> If using docker as above you can make this directory accessible to docker to make wallets and accounts persistent.  For example, for linux you could use the following command to list your wallets on Linux:
>
> ```
> docker run -v $HOME/.config/ethereum2/wallets:/data ethdo --basedir=/data wallet list
> ```
>
> This will allow you to use `ethdo` with or without docker, with the same location for wallets and accounts.

All ethdo comands take the following parameters:

  - `store`: the name of the storage system for wallets.  This can be one of "filesystem" (for local storage of the wallet) or "s3" (for remote storage of the wallet on [Amazon's S3](https://aws.amazon.com/s3/) storage system), and defaults to "filesystem"
  - `storepassphrase`: the passphrase for the store.  If this is empty the store is unencrypted
  - `walletpassphrase`: the passphrase for the wallet.  This is required for some wallet-centric operations such as creating new accounts
  - `passphrase`: the passphrase for the account.  This is required for some account-centric operations such as signing data

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

## Passphrase strength

`ethdo` will by default not allow creation or export of accounts or wallets with weak passphrases.  If a weak pasphrase is used then `ethdo` will refuse to continue.

If a weak passphrase is required, `ethdo` can be supplied with the `--allow-weak-passphrases` option which will force it to accept any passphrase, even if it is considered weak.

## Rules for account passphrases

Account passphrases are used in various places in `ethdo`.  Where they are used, the following rules apply:

  - commands that require passphrases to operate, for example unlocking an account, can be supplied with multiple passphrases.  If they are, then each passphrase is tried until one succeeds or they all fail
  - commands that require passphrases to create, for example creating an account, must be supplied with a single passphrase.  If more than one passphrase is supplied the command will fail

In addition, the following rules apply to passphrases supplied on the command line:

  - passphrases **must not** start with `0x`
  - passphrases **must not** contain the comma (,) character

# Commands

Command information, along with sample outputs and optional arguments, is available in [the usage section](https://github.com/wealdtech/ethdo/blob/master/docs/usage.md).

# HOWTO

There is a [HOWTO](https://github.com/wealdtech/ethdo/blob/master/docs/howto.md) that covers details about how to carry out various common tasks.  There is also a specific document that provides details of how to carry out [common conversions](docs/conversions.md) from mnemonic, to account, to deposit data, for launchpad-related configurations.

## Maintainers

Jim McDonald: [@mcdee](https://github.com/mcdee).

Special thanks to [@SuburbanDad](https://github.com/SuburbanDad) for updating xgo to allow for cross-compilation of `ethdo` releaes.

## Contribute

Contributions welcome. Please check out [the issues](https://github.com/wealdtech/ethdo/issues).

## License

[Apache-2.0](LICENSE) © 2019, 2020 Weald Technology Trading Ltd

