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
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
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

var err error

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

// cmdPath recurses up the command information to create a path for this command through commands and subcommands
func cmdPath(cmd *cobra.Command) string {
	if cmd.Parent() == nil || cmd.Parent().Name() == "ethdo" {
		return cmd.Name()
	}
	return fmt.Sprintf("%s:%s", cmdPath(cmd.Parent()), cmd.Name())
}

// setupLogging sets up the logging for commands that wish to write output
func setupLogging() {
	logFile := viper.GetString("log")
	if logFile == "" {
		home, err := homedir.Dir()
		errCheck(err, "Failed to access home directory")
		logFile = filepath.FromSlash(home + "/ethdo.log")
	}
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	errCheck(err, "Failed to open log file")
	log.SetOutput(f)
	log.SetFormatter(&log.JSONFormatter{})
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
	viper.BindPFlag("log", RootCmd.PersistentFlags().Lookup("log"))
	RootCmd.PersistentFlags().String("store", "filesystem", "Store for accounts")
	viper.BindPFlag("store", RootCmd.PersistentFlags().Lookup("store"))
	RootCmd.PersistentFlags().String("account", "", "Account name (in format \"wallet/account\")")
	viper.BindPFlag("account", RootCmd.PersistentFlags().Lookup("account"))
	RootCmd.PersistentFlags().String("storepassphrase", "", "Passphrase for store (if applicable)")
	viper.BindPFlag("storepassphrase", RootCmd.PersistentFlags().Lookup("storepassphrase"))
	RootCmd.PersistentFlags().String("walletpassphrase", "", "Passphrase for wallet (if applicable)")
	viper.BindPFlag("walletpassphrase", RootCmd.PersistentFlags().Lookup("walletpassphrase"))
	RootCmd.PersistentFlags().String("passphrase", "", "Passphrase for account (if applicable)")
	viper.BindPFlag("passphrase", RootCmd.PersistentFlags().Lookup("passphrase"))
	RootCmd.PersistentFlags().Bool("quiet", false, "do not generate any output")
	viper.BindPFlag("quiet", RootCmd.PersistentFlags().Lookup("quiet"))
	RootCmd.PersistentFlags().Bool("verbose", false, "generate additional output where appropriate")
	viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))
	RootCmd.PersistentFlags().Bool("debug", false, "generate debug output")
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
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
	viper.ReadInConfig()
}

//
// Helpers
//

// Add flags for commands that carry out transactions
func addTransactionFlags(cmd *cobra.Command, explanation string) {
	//	cmd.Flags().String("passphrase", "", fmt.Sprintf("passphrase for %s", explanation))
	//	cmd.Flags().String("privatekey", "", fmt.Sprintf("private key for %s", explanation))
	//	cmd.Flags().String("gasprice", "", "Gas price for the transaction")
	//	cmd.Flags().String("value", "", "Ether to send with the transaction")
	//	cmd.Flags().Int64("gaslimit", 0, "Gas limit for the transaction; 0 is auto-select")
	//	cmd.Flags().Int64("nonce", -1, "Nonce for the transaction; -1 is auto-select")
	//	cmd.Flags().Bool("wait", false, "wait for the transaction to be mined before returning")
	//	cmd.Flags().Duration("limit", 0, "maximum time to wait for transaction to complete before failing (default forever)")
}

func outputIf(condition bool, msg string) {
	if condition {
		fmt.Println(msg)
	}
}

func localContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
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

func verify(path string, data []byte, domain uint64, signature types.Signature) (bool, error) {
	account, err := accountFromPath(path)
	if err != nil {
		return false, err
	}
	return signature.Verify(data, account.PublicKey(), domain), nil
}
