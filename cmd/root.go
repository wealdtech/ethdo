// Copyright © 2019 Weald Technology Trading
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
	"errors"
	"fmt"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	types "github.com/wealdtech/go-eth2-types"

	wallet "github.com/wealdtech/go-eth2-wallet"
	wtypes "github.com/wealdtech/go-eth2-wallet-types"
)

var cfgFile string
var quiet bool
var verbose bool
var debug bool

// Root variables, present for all commands
var rootStore string
var rootAccount string
var rootStorePassphrase string
var rootWalletPassphrase string
var rootAccountPassphrase string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:              "ethdo",
	Short:            "Ethereum 2 CLI",
	Long:             `Manage common Ethereum 2 tasks from the command line.`,
	PersistentPreRun: persistentPreRun,
}

func persistentPreRun(cmd *cobra.Command, args []string) {
	if cmd.Name() == "help" {
		// User just wants help
		return
	}

	if cmd.Name() == "version" {
		// User just wants the version
		return
	}

	// We bind viper here so that we bind to the correct command
	quiet = viper.GetBool("quiet")
	verbose = viper.GetBool("verbose")
	debug = viper.GetBool("debug")
	rootStore = viper.GetString("store")
	rootAccount = viper.GetString("account")
	rootStorePassphrase = viper.GetString("storepassphrase")
	rootWalletPassphrase = viper.GetString("walletpassphrase")
	rootAccountPassphrase = viper.GetString("passphrase")

	if quiet && verbose {
		die("Cannot supply both quiet and verbose flags")
	}
	if quiet && debug {
		die("Cannot supply both quiet and debug flags")
	}

	// Set up our wallet store
	err := wallet.SetStore(rootStore, []byte(rootStorePassphrase))
	errCheck(err, "Failed to set up wallet store")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(_exit_failure)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ethdo.yaml)")
	RootCmd.PersistentFlags().String("log", "", "log activity to the named file (default $HOME/ethdo.log).  Logs are written for every action that generates a transaction")
	if err := viper.BindPFlag("log", RootCmd.PersistentFlags().Lookup("log")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("store", "filesystem", "Store for accounts")
	if err := viper.BindPFlag("store", RootCmd.PersistentFlags().Lookup("store")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("account", "", "Account name (in format \"wallet/account\")")
	if err := viper.BindPFlag("account", RootCmd.PersistentFlags().Lookup("account")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("storepassphrase", "", "Passphrase for store (if applicable)")
	if err := viper.BindPFlag("storepassphrase", RootCmd.PersistentFlags().Lookup("storepassphrase")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("walletpassphrase", "", "Passphrase for wallet (if applicable)")
	if err := viper.BindPFlag("walletpassphrase", RootCmd.PersistentFlags().Lookup("walletpassphrase")); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().String("passphrase", "", "Passphrase for account (if applicable)")
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
	RootCmd.PersistentFlags().Bool("debug", false, "generate debug output")
	if err := viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug")); err != nil {
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
		if err != nil {
			fmt.Println(err)
			os.Exit(_exit_failure)
		}

		// Search config in home directory with name ".ethdo" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".ethdo")
	}

	viper.SetEnvPrefix("ETHDO")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		// Don't report lack of config file...
		if !strings.Contains(err.Error(), "Not Found") {
			fmt.Println(err)
			os.Exit(_exit_failure)
		}
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

// walletAndAccountNamesFromPath breaks a path in to wallet and account names.
func walletAndAccountNamesFromPath(path string) (string, string, error) {
	if len(path) == 0 {
		return "", "", errors.New("invalid account format")
	}
	index := strings.Index(path, "/")
	if index == -1 {
		// Just the wallet
		return path, "", nil
	}
	if index == len(path)-1 {
		// Trailing /
		return path[:index], "", nil
	}
	return path[:index], path[index+1:], nil
}

// walletFromPath obtains a wallet given a path specification.
func walletFromPath(path string) (wtypes.Wallet, error) {
	walletName, _, err := walletAndAccountNamesFromPath(path)
	if err != nil {
		return nil, err
	}
	w, err := wallet.OpenWallet(walletName)
	if err != nil {
		if strings.Contains(err.Error(), "failed to decrypt wallet") {
			return nil, errors.New("Incorrect store passphrase")
		}
		return nil, err
	}
	return w, nil
}

// accountFromPath obtains an account given a path specification.
func accountFromPath(path string) (wtypes.Account, error) {
	wallet, err := walletFromPath(path)
	if err != nil {
		return nil, err
	}
	_, accountName, err := walletAndAccountNamesFromPath(path)
	if err != nil {
		return nil, err
	}
	if accountName == "" {
		return nil, errors.New("no account name")
	}

	if wallet.Type() == "hierarchical deterministic" && strings.HasPrefix(accountName, "m/") && rootWalletPassphrase != "" {
		err = wallet.Unlock([]byte(rootWalletPassphrase))
		if err != nil {
			return nil, errors.New("invalid wallet passphrase")
		}
		defer wallet.Lock()
	}
	return wallet.AccountByName(accountName)
}

func sign(path string, data []byte, domain uint64) (types.Signature, error) {
	assert(rootAccountPassphrase != "", "--passphrase is required")

	account, err := accountFromPath(path)
	if err != nil {
		return nil, err
	}
	err = account.Unlock([]byte(rootAccountPassphrase))
	if err != nil {
		return nil, err
	}
	defer account.Lock()
	return account.Sign(data, domain)
}
