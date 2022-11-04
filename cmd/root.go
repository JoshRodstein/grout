/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "grout",
	Short: GROUT,
	Long: GROUT + `Welcome to the Git Remote Migration Utility

This tool is designed to make finding and changing git remotes easier and less confusing.
This command by itself will run grout in interactive mode.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running interactive mode.")
		IAmGrout()
		fmt.Println("---------------------")

		targetRemoteURL = promptForInput(fmt.Sprintf("Target Remote Hostname (%s): ",
			defaultTargetHostname), defaultTargetHostname)
		newRemoteURL = promptForInput(fmt.Sprintf("New Remote Hostname (%s): ",
			defaultNewHostname), defaultNewHostname)
		targetOrganization = promptForInput(fmt.Sprintf("Target Username/Org (%s): ", "Target all if not set"), "")
		newOrganization = promptForInput(fmt.Sprintf("New Username/Org (%s): ", "Unchanged if not set"), newOrganization)
		targetDir = promptForInput(fmt.Sprintf("Target local directory (%s): ",
			"Defaults to './' if not set"), targetDir)

		// clean validate parameters
		newRemoteURL = trimSlashSuffix(newRemoteURL)
		targetRemoteURL = trimSlashSuffix(targetRemoteURL)
		newOrganization = strings.ReplaceAll(newOrganization, "/", "")
		targetOrganization = strings.ReplaceAll(targetOrganization, "/", "")
		if err := verifyTargetDirIsAbs(); err != nil {
			os.Exit(1)
		}

		// Prompt for confirmation of entered values
		confirmation := ParametersConfirmationOutput()
		input := promptForInput(confirmation, "")
		fmt.Println()
		if strings.Compare(strings.ToLower(input), Yes) != 0 {
			fmt.Println("Aborting changes")
			os.Exit(0)
		}

		fmt.Println("Generating plan...")

		// Search for git repos
		err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
			err = mapRepository(path, info, err, &repoMap)
			if err == filepath.SkipDir {
				return filepath.SkipDir
			}
			if err != nil {
				log.Fatalf("Unexpected error while walking directory tree: %s\n", err)
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}

		// calculate changes for repos in repoMap
		changeSet = createChangeSetFromMap(repoMap)
		writeChangeSetToFile(changeSet, defaultPlanFile)
		fmt.Printf("A change plan has been generated and is shown below. These changes have been saved to %s\n\n",
			defaultPlanFile)
		for _, plan := range changeSet.Plans {
			DisplayChangePlanForDirectory(plan)
		}
		DisplayBundledErrorsPlan()

		// Clear bundled errors from plan
		errorBundle.Count = 0
		errorBundle.Errors = nil

		// Display intent of plan and prompt for confirmation before proceeding
		if changeSet.Count > 0 {
			DisplayChangeIntention(changeSet)
			fmt.Println("---------------------")
			input := promptForInput("Enter '"+Yes+"' to accept and apply these changes: ", "")
			if strings.Compare(strings.ToLower(input), Yes) == 0 {
				err = executeChanges(changeSet)
				if err != nil {
					fmt.Println("grout was unable to execute the changes")
					os.Exit(1)
				}
				DisplayBundledErrorsUpdate()
				DisplayChangeResult(changeSet)
			} else {
				fmt.Println()
				fmt.Println("Aborting changes")
				os.Exit(0)
			}
		} else {
			DisplayBundledErrorsUpdate()
			fmt.Println("No Changes found.")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	currentDir, err = filepath.Abs(currentDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	targetDir = currentDir
	remoteType = defaultRemoteType

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.grout_bin.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output for logging/debugging ")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
			os.Exit(1)
		}

		// Search config in home directory with name ".grout_bin" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".grout_bin")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
