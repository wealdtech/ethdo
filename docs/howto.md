# How to achieve common tasks with ethdo

## Find out what ethdo can do

To find a list of topics that ethdo can carry out with the `ethdo help` command.

If you want more detailed information about the commands in a topic, they can be seen with the `ethdo help <topic>` command, for example:

```sh
ethdo help wallet
```

## List my wallets

The wallets you can currently access can be seen with the `ethdo wallet list` command.

## Create a new wallet

New wallets can be created with the `ethdo wallet create` command.  Each wallet has to have a unique name, for example:

```sh
ethdo wallet create --wallet="My wallet"
```

Additional options are available to decide the type of wallet and encryption.

## Create an HD wallet from an existing mnemonic

HD wallets can be created from an existing mnemonic by adding the `--mnemonic` parameter to `ethdo wallet create`, for example:

```sh
ethdo wallet create --wallet="Recreated wallet" --type=hd --wallet-passphrase="secret" --mnemonic="tooth moon mad fun romance athlete envelope next mix divert tip top symbol resemble stock family melody desk sheriff drift bargain need jaguar method"
```

## Back up a wallet

A wallet can be backed up with the `ethdo wallet export` command.  This creates an encrypted backup of the wallet, for example:

```sh
ethdo wallet export --wallet="My wallet" --passphrase="export secret" >export.dat
```

Note that by default the wallet backup is printed to the console, hence the `>export.dat` to redirect it to a file.

## Restore a wallet

A backed up wallet can be restored with the `ethdo wallet import` command, for example:

```sh
ethdo wallet import --data=export.dat --passphrase="export secret"
```

In this example the wallet to be imported is being read from the `export.dat` file.

Note that if a wallet with the same name already exists it cannot be imported.

## Where is my wallet?

Details of the location of a wallet can be found with the `ethdo wallet info` command, for example:

```sh
ethdo wallet info --verbose --wallet="My wallet"
```

This will provide, amongst other information, a `Location` line giving the directory where the wallet information resides.

## Recreate launchpad wallet and accounts

Recreating launchpad accounts requires two steps: recreating the wallet, and recreating the individual accounts.  All that is required is the mnemonic from the launchpad process.

To recreate the wallet with the given mnemonic run the following command (changing the wallet name, passphrase and mnemonic as required):

```sh
ethdo wallet create --wallet="Launchpad" --type=hd --wallet-passphrase=walletsecret --mnemonic="faculty key lamp panel appear choose express off absent dance strike twenty elephant expect swift that resist bicycle kind sun favorite evoke engage thumb"
```

Launchpad accounts are identified by their path.  The path can be seen in the filename of the keystore, for example the filename `keystore-m_12381_3600_1_0_0-1596891358.json` relates to a path of `m/12381/3600/1/0/0`.  It is also present directly in the keystore under the `path` key.

To create an account corresponding to this key with the account name "Account 1" you would use the command:

```sh
ethdo account create --account="Launchpad/Account 1" --wallet-passphrase=walletsecret --passphrase=secret --path=m/12381/3600/1/0/0
```
