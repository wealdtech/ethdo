# Changing withdrawal credentials.
When creating a validator it is possible to set its withdrawal credentials to those based upon a BLS private key (known as BLS withdrawal credentials, or "type 0" withdrawal credentials) or based upon an Ethereum execution address (known as execution withdrawal credentials, or "type 1" withdrawal credentials).  With the advent of the Capella hard fork, it is possible for rewards accrued on the consensus chain (also known as the beacon chain) to be sent to the execution chain.  However, for this to occur the validator's withdrawal credentials must be type 1.  Capella also brings a mechanism to change existing type 0 withdrawal credentials to type 1 withdrawal credentials, and this document outlines the process to change withdrawal credentials from type 0 to type 1 so that consensus rewards can be accessed.

**Once a validator has Ethereum execution credentials set they cannot be changed.  Please be careful when following this or any similar process to ensure you end up with the ability to access the rewards that will be sent to the execution address within the credentials.**

## Preparing for the process
A number of steps need to be taken to prepare for creating and broadcasting the credentials change operation.

### Accessing the beacon node
`ethdo` requires access to the beacon node at various points during the operation.  `ethdo` will attempt to find a local beacon node automatically, but if not then an explicit connection value will be required.  To find out if `ethdo` has access to the beacon node run:

```
ethdo node info --verbose
```

The result should be something similar to the following:

```
Version: teku/v22.9.1/linux-x86_64/-privatebuild-openjdk64bitservervm-java-14
Syncing: false
```

It is important to confirm that the "Syncing" value is "false".  If this is "true" it means that the node is currently syncing, and you will need to wait for the process to finish before proceeding.

If this command instead returns an error you will need to add an explicit connection string.  For example, if your beacon node is serving its REST API on port 12345 then you should add `--connection=http://localhost:12345` to all `ethdo` commands in this process, for example:

```sh
ethdo --connection=http://localhost:12345 node info --verbose
```

Note that some beacon nodes may require configuration to serve their REST API.  Please refer to the documentation of your specific beacon node to enable this.

### Validator reference
There are three options to reference a validator:

- the `ethdo` account of the validator (in format wallet/account)
- the validator's public key (in format 0x…)
- the validator's index (in format 123…)

Any of these can be passed to the following commands with the `--validator` parameter.  You need to ensure that you have this information before starting the process.

**In the following examples we will use the validator with index 123.  Please replace this with the reference to your validator in all commands.**

### Execution address
The execution address will be the address to which all Ether held by the validator from the consensus chain will be sent.  It is important to understand that at time of writing this value cannot be changed, so it is critical that one of the following criteria are met:

- the private keys for the Ethereum address are known
- the Ethereum address is secured by a hardware wallet
- the Ethereum address is that of a smart contract with the ability to withdraw funds

The execution address must be supplied in [EIP-55](https://eips.ethereum.org/EIPS/eip-55) format, _i.e._ using mixed case for checksum.  An example of a mixed-case Ethereum address is `0x8f0844Fd51E31ff6Bf5baBe21DCcf7328E19Fd9F`

**In the following examples we will use an execution address of 0x8f…9F.  Please replace this with the your execution address in all commands.**

### Online or offline
It is possible to generate the withdrawal credentials change operation either online or offline.

In _online_ mode the credentials will be generated on a server that has both access to the internet and access to the private keys of the existing withdrawal credentials.  This is the easiest process, however due to it involving private keys on a computer connected to the internet some consider this insecure.

In _offline_ mode there are two servers: one with access to the internet, and one with access to the private keys.  This is the most secure process, however requires additional steps to accomplish.

It is a personal choice as to if an online or offline method is chosen to generate the credentials change operation.  Instructions for both methods are present.

## The process
### Check your current validator credentials
The first step will be to confirm that the validator can be found on-chain.  To do so, run the following command:

```sh
ethdo validator credentials get --validator=123
```

This should return information similar to the following:

```
BLS credentials: 0x00ebf119d469a31ff2a534d176e6d594046a2367f7a36848009f70f3cb9a9dd1
```

This result should start with the phrase "BLS credentials", which means that these credentials must be upgraded to an Ethereum execution address to receive withdrawals.  If instead the result starts with the phrase "Ethereum execution address" it means that the credentials are already set to an Ethereum execution address and no further action is necessary (or possible).


The next step depends on if you have access to your keys 

### Generate and publish the credentials change operation (online)
The steps for generating and publishing the credentials change operation online depend on the method by which you access your current withdrawal key.

#### Using a mnemonic
Many stakers will have generated their validators from a mnemonic.  A mnemonic is a 24-word phrase from which withdrawal and validator keys are derived using a path.

- m/12381/3600/_i_/0 is the path to a withdrawal key, where _i_ starts at 0 for the first validator, 1 for the second validator, _etc._
- m/12381/3600/_i_/0/0 is the path to a validator key, where _i_ starts at 0 for the first validator, 1 for the second validator, _etc._

The first step will be to confirm that the mnemonic provides the appropriate validator key.  To do so run:

```
ethdo account derive --mnemonic='abandon … art' --path='m/12381/3600/0/0/0'
```

replacing the first '0' in the path with the validator number (remember that numbering starts at 0 for the first validator).  This will provide an output similar to:

```
Public key: 0xb384f767d964e100c8a9b21018d08c25ffebae268b3ab6d610353897541971726dbfc3c7463884c68a531515aab94c87
```

The displayed public key should match the public key of the validator of which you are attempting to change the credentials.  If not, then do not proceed further and obtain help to understand why there is a mismatch.

Assuming the displayed public key does match the public key of the validator the next step is to confirm the current withdrawal credentials.  To do su run:

```
ethdo account derive --mnemonic='abandon … art' --path='m/12381/0/0' --show-withdrawal-credentials
```

again replacing the first '0' in the path with the validator number.  This will provide an output similar to:

```
Public key: 0x99b1f1d84d76185466d86c34bde1101316afddae76217aa86cd066979b19858c2c9d9e56eebc1e067ac54277a61790db
Withdrawal credentials: 0x008ba1cc4b091b91c1202bba3f508075d6ff565c77e559f0803c0792e0302bf1
```

The displayed withdrawal credentials should match the current withdrawal credentials of your validator (note that these were obtained in an earlier step so you can use the output there to confirm that they match).  If not, then do not proceed further and obtain help to understand why there is a mismatch.

Once you are comfortable that the mnemonic and path provide the correct result you can generate and broadcast the credentials change operation with the following command:

```
ethdo validator credentials set --validator=123 --execution-address=0x8f…9F --mnemonic='abandon … art' --path='m/12381/0/0'
```

again replacing the first '0' in the path with the validator number, and using your own execution address as explained earlier in the guide.

#### Using a private key
If you have the private key from which the current withdrawal credentials were derived this can be used to generate and broadcast the credentials change operation with the following command:

```
ethdo validator credentials set --validator=123 --execution-address=0x8f…9F --private-key=0x3b…9c
```

using your own execution address as explained earlier in the guide, and your own private key.

#### Using an account
If you used `ethdo` to generate your validator deposit data you will likely have used a separate account to generate the withdrawal credentials.  You can specify the account to generate and broadcast the credentials change operation with the following command:

```
ethdo validator credentials set --validator=123 --execution-address=0x8f…9F --account=Wallet/Account --passphrase=secret
```

setting the execution address, account and passphrase to your own values.

### Generate the credentials change operation (offline)
Generating the credentials change operation offline requires information from the online component, so is somewhat more involved than the online process, however does not expose mnemonics, private keys, or passphrases to servers that are connected to the internet.  The process is below.

#### Obtain data required for offline generation.
Generating the credentials change operation requires information that comes from an online beacon node.  As such, on your _online_ server you need to run the following command:

```
ethdo chain info --prepare-offline
```

This will return something similar to the following response:

```
Add the following to your command to run it offline:
  --offline --genesis-validators=root=0x043db0d9a83813551ee2f33450d23797757d430911a9320530ad8a0eabc43efb --fork-version=0x03001020
```

This information needs to be copied to your offline server to continue.

#### Generate signed operation
Generating the signed operation offline

#### Using a mnemonic
Many stakers will have generated their validators from a mnemonic.  A mnemonic is a 24-word phrase from which withdrawal and validator keys are derived using a path.

- m/12381/3600/_i_/0 is the path to the _i_th withdrawal key, where _i_ starts at 0 for the first validator, 1 for the second validator, _etc._
- m/12381/3600/_i_/0/0 is the path to the _i_th validator key, where _i_ starts at 0 for the first validator, 1 for the second validator, _etc._

The first step will be to confirm that the mnemonic provides the appropriate validator key.  To do so run:

```
ethdo account derive --mnemonic='abandon … art' --path='m/12381/3600/0/0/0'
```

replacing the first '0' in the path with the validator number (remember that numbering starts at 0 for the first validator).  This will provide an output similar to:

```
Public key: 0xb384f767d964e100c8a9b21018d08c25ffebae268b3ab6d610353897541971726dbfc3c7463884c68a531515aab94c87
```

The displayed public key should match the public key of the validator of which you are attempting to change the credentials.  If not, then do not proceed further and obtain help to understand why there is a mismatch.

Assuming the displayed public key does match the public key of the validator the next step is to confirm the current withdrawal credentials.  To do su run:

```
ethdo account derive --mnemonic='abandon … art' --path='m/12381/0/0' --show-withdrawal-credentials
```

again replacing the first '0' in the path with the validator number.  This will provide an output similar to:

```
Public key: 0x99b1f1d84d76185466d86c34bde1101316afddae76217aa86cd066979b19858c2c9d9e56eebc1e067ac54277a61790db
Withdrawal credentials: 0x008ba1cc4b091b91c1202bba3f508075d6ff565c77e559f0803c0792e0302bf1
```

The displayed withdrawal credentials should match the current withdrawal credentials of your validator (note that these were obtained in an earlier step so you can use the output there to confirm that they match).  If not, then do not proceed further and obtain help to understand why there is a mismatch.

Once you are comfortable that the mnemonic and path provide the correct result you can generate the credentials change operation with the following command:

```
ethdo validator credentials set --offline --genesis-validators=root=0x04…fb --fork-version=0x03…20 --validator=123 --execution-address=0x8f…9F --mnemonic='abandon … art' --path='m/12381/0/0'
```

again replacing the first '0' in the path with the validator number, and using your own execution address as explained earlier in the guide.  This will produce output similar to the following:

```
{"message":{"validator_index":"123","from_bls_pubkey":"0xad1868210a0cff7aff22633c003c503d4c199c8dcca13bba5b3232fc784d39d3855936e94ce184c3ce27bf15d4347695","to_execution_address":"0x388ea662ef2c223ec0b047d41bf3c0f362142ad5"},"signature":"0x8fcc8ceb75cbea891540150efc7df3e482a74592f89f3fc62a2d034381c776fcd42faad82af7a4af7fb84168a74981ce0ec96cf059e134eaa979c67425138f1915d1a8b1b6056401a9f7a2e79ed673f4b0c6b6ae1f60cff5996318e4769d0642"}

```

#### Using a private key
If you have the private key from which the current withdrawal credentials were derived this can be used to generate the credentials change operation with the following command:

```
ethdo validator credentials set --offline --genesis-validators=root=0x04…fb --fork-version=0x03…20 --validator=123 --execution-address=0x8f…9F --private-key=0x3b…9c
```

using your own execution address as explained earlier in the guide, and your own private key.  This will produce output similar to the following:

```
{"message":{"validator_index":"123","from_bls_pubkey":"0xad1868210a0cff7aff22633c003c503d4c199c8dcca13bba5b3232fc784d39d3855936e94ce184c3ce27bf15d4347695","to_execution_address":"0x388ea662ef2c223ec0b047d41bf3c0f362142ad5"},"signature":"0x8fcc8ceb75cbea891540150efc7df3e482a74592f89f3fc62a2d034381c776fcd42faad82af7a4af7fb84168a74981ce0ec96cf059e134eaa979c67425138f1915d1a8b1b6056401a9f7a2e79ed673f4b0c6b6ae1f60cff5996318e4769d0642"}
```

#### Using an account
If you used `ethdo` to generate your validator deposit data you will likely have used a separate account to generate the withdrawal credentials.  You can specify the account to generate the credentials change operation with the following command:

```
ethdo validator credentials set --offline --genesis-validators=root=0x04…fb --fork-version=0x03…20 --validator=123 --execution-address=0x8f…9F --account=Wallet/Account --passphrase=secret
```

setting the execution address, account and passphrase to your own values.  This will produce output similar to the following:

```
{"message":{"validator_index":"123","from_bls_pubkey":"0xad1868210a0cff7aff22633c003c503d4c199c8dcca13bba5b3232fc784d39d3855936e94ce184c3ce27bf15d4347695","to_execution_address":"0x388ea662ef2c223ec0b047d41bf3c0f362142ad5"},"signature":"0x8fcc8ceb75cbea891540150efc7df3e482a74592f89f3fc62a2d034381c776fcd42faad82af7a4af7fb84168a74981ce0ec96cf059e134eaa979c67425138f1915d1a8b1b6056401a9f7a2e79ed673f4b0c6b6ae1f60cff5996318e4769d0642"}
```

### Broadcasting a previously-generated credentials change operation
An online server can broadcast the result of the previous step.  Note that the data does not expose any sensitive information such as private keys, and as such is safe to be accessed by the online server.  Broadcasting the operation is a simple case of supplying it to `ethdo`:

```
ethdo validator credentials set --signed-operation='{"message":{"validator_index":"123","from_bls_pubkey":"0xad1868210a0cff7aff22633c003c503d4c199c8dcca13bba5b3232fc784d39d3855936e94ce184c3ce27bf15d4347695","to_execution_address":"0x388ea662ef2c223ec0b047d41bf3c0f362142ad5"},"signature":"0x8fcc8ceb75cbea891540150efc7df3e482a74592f89f3fc62a2d034381c776fcd42faad82af7a4af7fb84168a74981ce0ec96cf059e134eaa979c67425138f1915d1a8b1b6056401a9f7a2e79ed673f4b0c6b6ae1f60cff5996318e4769d0642"}'
```

Alternatively, if the operation is stored on a filesystem, for example on a USB device from where it was copied from the offline server, it can be used with:

```
ethdo validator credentials set --signed-operation=/path/to/signed/operation
```

## Confirming the process has succeeded
The final step is confirming the operation has taken place.  To do so, run the following command on an online server:

```sh
ethdo validator credentials get --validator=123
```

The result should start with the phrase "Ethereum execution address" and display the execution address you chose at the beginning of the process, for example:

```
Ethereum execution address: 0x8f0844Fd51E31ff6Bf5baBe21DCcf7328E19Fd9F
```

If the result starts with the phrase "BLS credentials" then it may be that the operation has yet to be incorporated on the chain, please wait a few minutes and check again.  If this continues to be the case please obtain help to understand why the change operation failed to work.
