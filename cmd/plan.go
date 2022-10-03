/*
Copyright © 2021 Joshua Rodstein joshuarodstein@gmail.com

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

	"github.com/spf13/cobra"
)

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Generate a plan for updating your git remotes",
	Long: `
Generate a plan for updating git remotes:

  Search all folders under given directory for git repositories and 
  Spit out a plan for updating git remotes to a new URL. Plan is saved 
  as test-grout-plan.json unless otherwise specified.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := verifyTargetDirIsAbs(); err != nil {
			os.Exit(1)
		}

		fmt.Println()
		fmt.Println("Plan parameters:")
		confirmation := ParametersConfirmationOutput()
		input := promptForInput(confirmation, "")
		fmt.Println()
		if strings.Compare(strings.ToLower(input), Yes) != 0 {
			fmt.Println("Aborting changes")
			os.Exit(0)
		}

		fmt.Println("Generating plan...")

		// Walk directory tree and map repositories
		err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
			err = mapRepository(path, info, err, &repoMap)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}

		changeSet = createChangeSetFromMap(repoMap)
		writeChangeSetToFile(changeSet, defaultPlanFile)

		if changeSet.Count > 0 {
			fmt.Printf("A change plan has been generated and is shown below. These changes have been saved to %s\n\n",
				defaultPlanFile)
			for _, plan := range changeSet.Plans {
				DisplayChangePlanForDirectory(plan)
			}
			DisplayBundledErrorsPlan()
			DisplayChangeCount(changeSet)
		} else {
			DisplayBundledErrorsPlan()
			fmt.Println("\nNo Changes found.")
		}

	},
}

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.Flags().StringVarP(&targetDir, "directory", "d", targetDir, "Set search directory")
	planCmd.Flags().StringVar(&targetRemoteURL, "find-url", defaultTargetHostname, "set remote url for remote update")
	planCmd.Flags().StringVar(&newRemoteURL, "set-url", defaultNewHostname, "set target url for remote update")
	planCmd.Flags().StringVar(&targetOrganization, "find-org", "", "set target org for remote update")
	planCmd.Flags().StringVar(&newOrganization, "set-org", "", "set target org for remote update")
	planCmd.Flags().StringVar(&remoteType, "remote-type", defaultRemoteType, "set target org for remote update")
	planCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// clean and validate parameters
	newRemoteURL = trimSlashSuffix(newRemoteURL)
	targetRemoteURL = trimSlashSuffix(targetRemoteURL)
	newOrganization = strings.ReplaceAll(newOrganization, "/", "")
}
