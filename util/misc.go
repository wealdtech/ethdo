// Copyright © 2019, 2022 Weald Technology Trading
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	dirk "github.com/wealdtech/go-eth2-wallet-dirk"
	filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
	s3 "github.com/wealdtech/go-eth2-wallet-store-s3"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// SetupStore sets up the account store.
func SetupStore() error {
	var store e2wtypes.Store
	var err error
	if viper.GetString("remote") != "" {
		// We are using a remote account manager, so no local setup required.
		return nil
	}

	// Set up our wallet store.
	switch viper.GetString("store") {
	case "s3":
		if GetBaseDir() != "" {
			return errors.New("basedir does not apply to the s3 store")
		}
		store, err = s3.New(s3.WithPassphrase([]byte(GetStorePassphrase("s3"))),
			s3.WithID([]byte(viper.GetString("stores.s3.id"))),
			s3.WithEndpoint(viper.GetString("stores.s3.endpoint")),
			s3.WithRegion(viper.GetString("stores.s3.region")),
			s3.WithBucket(viper.GetString("stores.s3.bucket")),
			s3.WithPath(viper.GetString("stores.s3.path")),
			s3.WithCredentialsID(viper.GetString("stores.s3.credentials.id")),
			s3.WithCredentialsSecret(viper.GetString("stores.s3.credentials.secret")),
		)
		if err != nil {
			return errors.Wrap(err, "failed to access Amazon S3 wallet store")
		}
	case "filesystem":
		opts := make([]filesystem.Option, 0)
		if GetStorePassphrase("filesystem") != "" {
			opts = append(opts, filesystem.WithPassphrase([]byte(GetStorePassphrase("filesystem"))))
		}
		if GetBaseDir() != "" {
			opts = append(opts, filesystem.WithLocation(GetBaseDir()))
		}
		store = filesystem.New(opts...)
	default:
		return fmt.Errorf("unsupported wallet store %s", viper.GetString("store"))
	}
	if err := e2wallet.UseStore(store); err != nil {
		return errors.Wrap(err, "failed to use defined wallet store")
	}
	viper.Set("store", store)

	return nil
}

// WalletFromInput obtains a wallet given the information in the viper variable
// "account", or if not present the viper variable "wallet".
func WalletFromInput(ctx context.Context) (e2wtypes.Wallet, error) {
	switch {
	case viper.GetString("account") != "":
		return WalletFromPath(ctx, viper.GetString("account"))
	case viper.GetString("wallet") != "":
		return WalletFromPath(ctx, viper.GetString("wallet"))
	default:
		return nil, errors.New("cannot determine wallet")
	}
}

// WalletFromPath obtains a wallet given a path specification.
func WalletFromPath(ctx context.Context, path string) (e2wtypes.Wallet, error) {
	walletName, _, err := e2wallet.WalletAndAccountNames(path)
	if err != nil {
		return nil, err
	}
	if viper.GetString("remote") != "" {
		if viper.GetString("client-cert") == "" {
			return nil, errors.New("remote connections require client-cert")
		}
		if viper.GetString("client-key") == "" {
			return nil, errors.New("remote connections require client-key")
		}
		credentials, err := dirk.ComposeCredentials(ctx, viper.GetString("client-cert"), viper.GetString("client-key"), viper.GetString("server-ca-cert"))
		if err != nil {
			return nil, errors.Wrap(err, "failed to build dirk credentials")
		}

		endpoints, err := remotesToEndpoints([]string{viper.GetString("remote")})
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse remote servers")
		}

		return dirk.Open(ctx,
			dirk.WithName(walletName),
			dirk.WithCredentials(credentials),
			dirk.WithEndpoints(endpoints),
			dirk.WithTimeout(viper.GetDuration("timeout")),
		)
	}
	wallet, err := e2wallet.OpenWallet(walletName)
	if err != nil {
		if strings.Contains(err.Error(), "failed to decrypt wallet") {
			return nil, errors.New("Incorrect store passphrase")
		}
		return nil, err
	}
	return wallet, nil
}

// WalletAndAccountFromInput obtains the wallet and account given the information in the viper variable "account".
func WalletAndAccountFromInput(ctx context.Context) (e2wtypes.Wallet, e2wtypes.Account, error) {
	return WalletAndAccountFromPath(ctx, viper.GetString("account"))
}

// WalletAndAccountFromPath obtains the wallet and account given a path specification.
func WalletAndAccountFromPath(ctx context.Context, path string) (e2wtypes.Wallet, e2wtypes.Account, error) {
	wallet, err := WalletFromPath(ctx, path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to open wallet for account")
	}
	_, accountName, err := e2wallet.WalletAndAccountNames(path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to obtain account name")
	}
	if accountName == "" {
		return nil, nil, errors.New("no account name")
	}

	if wallet.Type() == "hierarchical deterministic" && strings.HasPrefix(accountName, "m/") {
		if GetWalletPassphrase() == "" {
			return nil, nil, errors.New("walletpassphrase is required for direct path derivations")
		}

		locker, isLocker := wallet.(e2wtypes.WalletLocker)
		if isLocker {
			err = locker.Unlock(ctx, []byte(GetWalletPassphrase()))
			if err != nil {
				return nil, nil, errors.New("failed to unlock wallet")
			}
			defer relockAccount(locker)
		}
	}

	accountByNameProvider, isAccountByNameProvider := wallet.(e2wtypes.WalletAccountByNameProvider)
	if !isAccountByNameProvider {
		return nil, nil, errors.New("wallet cannot obtain accounts by name")
	}
	account, err := accountByNameProvider.AccountByName(ctx, accountName)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to obtain account")
	}
	return wallet, account, nil
}

// WalletAndAccountsFromPath obtains the wallet and matching accounts given a path specification.
func WalletAndAccountsFromPath(ctx context.Context, path string) (e2wtypes.Wallet, []e2wtypes.Account, error) {
	wallet, err := WalletFromPath(ctx, path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to open wallet for account")
	}

	_, accountSpec, err := e2wallet.WalletAndAccountNames(path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to obtain account specification")
	}
	if accountSpec == "" {
		accountSpec = "^.*$"
	} else {
		accountSpec = fmt.Sprintf("^%s$", accountSpec)
	}
	re := regexp.MustCompile(accountSpec)

	accounts := make([]e2wtypes.Account, 0)
	for account := range wallet.Accounts(ctx) {
		if re.MatchString(account.Name()) {
			accounts = append(accounts, account)
		}
	}

	// Tidy up accounts by name.
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].Name() < accounts[j].Name()
	})

	return wallet, accounts, nil
}

// BestPublicKey returns the best public key for operations.
// It prefers the composite public key if present, otherwise the public key.
func BestPublicKey(account e2wtypes.Account) (e2types.PublicKey, error) {
	var pubKey e2types.PublicKey
	publicKeyProvider, isCompositePublicKeyProvider := account.(e2wtypes.AccountCompositePublicKeyProvider)
	if isCompositePublicKeyProvider {
		pubKey = publicKeyProvider.CompositePublicKey()
	} else {
		publicKeyProvider, isPublicKeyProvider := account.(e2wtypes.AccountPublicKeyProvider)
		if isPublicKeyProvider {
			pubKey = publicKeyProvider.PublicKey()
		} else {
			return nil, errors.New("account does not provide a public key")
		}
	}
	return pubKey, nil
}

// relockAccount locks an account; generally called as a defer after an account is unlocked so handles its own error.
func relockAccount(locker e2wtypes.AccountLocker) {
	if err := locker.Lock(context.Background()); err != nil {
		panic(err)
	}
}

// remotesToEndpoints generates endpoints from remote addresses.
func remotesToEndpoints(remotes []string) ([]*dirk.Endpoint, error) {
	endpoints := make([]*dirk.Endpoint, 0)
	for _, remote := range remotes {
		parts := strings.Split(remote, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid remote %q", remote)
		}
		port, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("invalid port in remote %q", remote))
		}
		endpoints = append(endpoints, dirk.NewEndpoint(parts[0], uint32(port)))
	}
	return endpoints, nil
}
