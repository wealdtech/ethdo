// Copyright © 2019 - 2023 Weald Technology Trading.
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

package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wealdtech/ethdo/util"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	dirk "github.com/wealdtech/go-eth2-wallet-dirk"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:               "ethdo",
	Short:             "Ethereum consensus layer CLI",
	Long:              `Manage common Ethereum consensus layer tasks from the command line.`,
	PersistentPreRunE: persistentPreRunE,
}

// bindings are the command-specific bindings.
var bindings = map[string]func(cmd *cobra.Command){
	"account/create":     accountCreateBindings,
	"account/derive":     accountDeriveBindings,
	"account/import":     accountImportBindings,
	"attester/duties":    attesterDutiesBindings,
	"attester/inclusion": attesterInclusionBindings,
	"block/analyze":      blockAnalyzeBindings,
	"block/info":         blockInfoBindings,
	"chain/eth1votes":    chainEth1VotesBindings,
	"chain/info":         chainInfoBindings,
	"chain/queues":       chainQueuesBindings,
	"chain/spec":         chainSpecBindings,
	"chain/time":         chainTimeBindings,
	"chain/verify/signedcontributionandproof": chainVerifySignedContributionAndProofBindings,
	"epoch/summary":             epochSummaryBindings,
	"exit/verify":               exitVerifyBindings,
	"node/events":               nodeEventsBindings,
	"proposer/duties":           proposerDutiesBindings,
	"slot/time":                 slotTimeBindings,
	"synccommittee/inclusion":   synccommitteeInclusionBindings,
	"synccommittee/members":     synccommitteeMembersBindings,
	"validator/credentials/get": validatorCredentialsGetBindings,
	"validator/credentials/set": validatorCredentialsSetBindings,
	"validator/depositdata":     validatorDepositdataBindings,
	"validator/duties":          validatorDutiesBindings,
	"validator/exit":            validatorExitBindings,
	"validator/info":            validatorInfoBindings,
	"validator/keycheck":        validatorKeycheckBindings,
	"validator/summary":         validatorSummaryBindings,
	"validator/yield":           validatorYieldBindings,
	"validator/expectation":     validatorExpectationBindings,
	"validator/withdrawal":      validatorWithdrawalBindings,
	"wallet/batch":              walletBatchBindings,
	"wallet/create":             walletCreateBindings,
	"wallet/import":             walletImportBindings,
	"wallet/sharedexport":       walletSharedExportBindings,
	"wallet/sharedimport":       walletSharedImportBindings,
}

func persistentPreRunE(cmd *cobra.Command, _ []string) error {
	if cmd.Name() == "help" {
		// User just wants help
		return nil
	}

	if cmd.Name() == "version" {
		// User just wants the version
		return nil
	}

	// Disable service logging.
	zerolog.SetGlobalLevel(zerolog.Disabled)

	// We bind viper here so that we bind to the correct command.
	quiet := viper.GetBool("quiet")
	verbose := viper.GetBool("verbose")
	debug := viper.GetBool("debug")

	// Command-specific bindings.
	if bindingsFunc, exists := bindings[commandPath(cmd)]; exists {
		bindingsFunc(cmd)
	}

	if quiet && verbose {
		fmt.Println("Cannot supply both quiet and verbose flags")
	}
	if quiet && debug {
		fmt.Println("Cannot supply both quiet and debug flags")
	}

	return util.SetupStore()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(_exitFailure)
	}
}

func init() {
	// Initialise our BLS library.
	if err := e2types.InitBLS(); err != nil {
		fmt.Println(err)
		os.Exit(_exitFailure)
	}

	cobra.OnInitialize(initConfig)
	addPersistentFlags()
}

func addPersistentFlags() {
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ethdo.yaml)")

	RootCmd.PersistentFlags().String("log", "", "log activity to the named file (default $HOME/ethdo.log).  Logs are written for every action that generates a transaction")
	if err := viper.BindPFlag("log", RootCmd.PersistentFlags().Lookup("log")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("store", "filesystem", "Store for accounts")
	if err := viper.BindPFlag("store", RootCmd.PersistentFlags().Lookup("store")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("account", "", `Account name (in format "<wallet>/<account>")`)
	if err := viper.BindPFlag("account", RootCmd.PersistentFlags().Lookup("account")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("mnemonic", "", "Mnemonic to provide access to an account")
	if err := viper.BindPFlag("mnemonic", RootCmd.PersistentFlags().Lookup("mnemonic")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("path", "", "Hierarchical derivation path used with mnemonic to provide access to an account")
	if err := viper.BindPFlag("path", RootCmd.PersistentFlags().Lookup("path")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("private-key", "", "Private key to provide access to an account or validator")
	if err := viper.BindPFlag("private-key", RootCmd.PersistentFlags().Lookup("private-key")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("public-key", "", "public key to provide access to an account")
	if err := viper.BindPFlag("public-key", RootCmd.PersistentFlags().Lookup("public-key")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("basedir", "", "Base directory for filesystem wallets")
	if err := viper.BindPFlag("basedir", RootCmd.PersistentFlags().Lookup("basedir")); err != nil {
		panic(err)
	}
	if err := RootCmd.PersistentFlags().MarkDeprecated("basedir", "use --base-dir"); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("base-dir", "", "Base directory for filesystem wallets")
	if err := viper.BindPFlag("base-dir", RootCmd.PersistentFlags().Lookup("base-dir")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("storepassphrase", "", "Passphrase for store (if applicable)")
	if err := viper.BindPFlag("storepassphrase", RootCmd.PersistentFlags().Lookup("storepassphrase")); err != nil {
		panic(err)
	}
	if err := RootCmd.PersistentFlags().MarkDeprecated("storepassphrase", "use --store-passphrase"); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("store-passphrase", "", "Passphrase for store (if applicable)")
	if err := viper.BindPFlag("store-passphrase", RootCmd.PersistentFlags().Lookup("store-passphrase")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("walletpassphrase", "", "Passphrase for wallet (if applicable)")
	if err := viper.BindPFlag("walletpassphrase", RootCmd.PersistentFlags().Lookup("walletpassphrase")); err != nil {
		panic(err)
	}
	if err := RootCmd.PersistentFlags().MarkDeprecated("walletpassphrase", "use --wallet-passphrase"); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("wallet-passphrase", "", "Passphrase for wallet (if applicable)")
	if err := viper.BindPFlag("wallet-passphrase", RootCmd.PersistentFlags().Lookup("wallet-passphrase")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().StringSlice("passphrase", nil, "Passphrase for account (if applicable)")
	if err := viper.BindPFlag("passphrase", RootCmd.PersistentFlags().Lookup("passphrase")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().Bool("quiet", false, "do not generate any output")
	if err := viper.BindPFlag("quiet", RootCmd.PersistentFlags().Lookup("quiet")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().Bool("verbose", false, "generate additional output where appropriate")
	if err := viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().Bool("json", false, "generate JSON output where available")
	if err := viper.BindPFlag("json", RootCmd.PersistentFlags().Lookup("json")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().Bool("debug", false, "generate debug output")
	if err := viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("connection", "", "URL to an Ethereum 2 node's REST API endpoint")
	if err := viper.BindPFlag("connection", RootCmd.PersistentFlags().Lookup("connection")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().Duration("timeout", 30*time.Second, "the time after which a network request will be considered failed.  Increase this if you are running on an error-prone, high-latency or low-bandwidth connection")
	if err := viper.BindPFlag("timeout", RootCmd.PersistentFlags().Lookup("timeout")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("remote", "", "connection to a remote wallet daemon")
	if err := viper.BindPFlag("remote", RootCmd.PersistentFlags().Lookup("remote")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("client-cert", "", "location of a client certificate file when connecting to the remote wallet daemon")
	if err := viper.BindPFlag("client-cert", RootCmd.PersistentFlags().Lookup("client-cert")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("client-key", "", "location of a client key file when connecting to the remote wallet daemon")
	if err := viper.BindPFlag("client-key", RootCmd.PersistentFlags().Lookup("client-key")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("server-ca-cert", "", "location of the server certificate authority certificate when connecting to the remote wallet daemon")
	if err := viper.BindPFlag("server-ca-cert", RootCmd.PersistentFlags().Lookup("server-ca-cert")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().Bool("allow-weak-passphrases", false, "allow passphrases that use common words, are short, or generally considered weak")
	if err := viper.BindPFlag("allow-weak-passphrases", RootCmd.PersistentFlags().Lookup("allow-weak-passphrases")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().Bool("allow-insecure-connections", false, "allow insecure connections to remote beacon nodes")
	if err := viper.BindPFlag("allow-insecure-connections", RootCmd.PersistentFlags().Lookup("allow-insecure-connections")); err != nil {
		panic(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		errCheck(err, "could not find home directory")

		// Search config in home directory with name ".ethdo" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".ethdo")
	}

	viper.SetEnvPrefix("ETHDO")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		// Don't report lack of config file...
		assert(strings.Contains(err.Error(), "Not Found"), "failed to read configuration")
	}
}

//
// Helpers
//

func outputIf(condition bool, msg string) {
	if condition {
		fmt.Println(msg)
	}
}

// walletFromInput obtains a wallet given the information in the viper variable
// "account", or if not present the viper variable "wallet".
func walletFromInput(ctx context.Context) (e2wtypes.Wallet, error) {
	if viper.GetString("account") != "" {
		return walletFromPath(ctx, viper.GetString("account"))
	}
	return walletFromPath(ctx, viper.GetString("wallet"))
}

// walletFromPath obtains a wallet given a path specification.
func walletFromPath(ctx context.Context, path string) (e2wtypes.Wallet, error) {
	walletName, _, err := e2wallet.WalletAndAccountNames(path)
	if err != nil {
		return nil, err
	}
	if viper.GetString("remote") != "" {
		assert(viper.GetString("client-cert") != "", "remote connections require client-cert")
		assert(viper.GetString("client-key") != "", "remote connections require client-key")
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

// walletAndAccountFromInput obtains the wallet and account given the information in the viper variable "account".
func walletAndAccountFromInput(ctx context.Context) (e2wtypes.Wallet, e2wtypes.Account, error) {
	return walletAndAccountFromPath(ctx, viper.GetString("account"))
}

// walletAndAccountFromPath obtains the wallet and account given a path specification.
func walletAndAccountFromPath(ctx context.Context, path string) (e2wtypes.Wallet, e2wtypes.Account, error) {
	wallet, err := walletFromPath(ctx, path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to open wallet for account")
	}
	_, accountName, err := e2wallet.WalletAndAccountNames(path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to obtain accout name")
	}
	if accountName == "" {
		return nil, nil, errors.New("no account name")
	}

	if wallet.Type() == "hierarchical deterministic" && strings.HasPrefix(accountName, "m/") {
		assert(getWalletPassphrase() != "", "--walletpassphrase is required for direct path derivations")

		locker, isLocker := wallet.(e2wtypes.WalletLocker)
		if isLocker {
			err = locker.Unlock(ctx, []byte(util.GetWalletPassphrase()))
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

// relockAccount locks an account; generally called as a defer after an account is unlocked.
func relockAccount(locker e2wtypes.AccountLocker) {
	errCheck(locker.Lock(context.Background()), "failed to re-lock account")
}

func commandPath(cmd *cobra.Command) string {
	path := ""
	for {
		path = fmt.Sprintf("%s/%s", cmd.Name(), path)
		if cmd.Parent().Name() == "ethdo" {
			return strings.TrimRight(path, "/")
		}
		cmd = cmd.Parent()
	}
}
