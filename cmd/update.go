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
	"strings"

	//"github.com/go-git/go-git/v5"
	"os"

	"github.com/spf13/cobra"
)

var planFile string

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Execute changes in a plan",
	Long: `
Execute changes in a plan:
	
  Load and execute a plan from file. grout looks for test-grout-plan.json in 
  it's current directory unless otherwise specified.`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("Loading plan...")
		changeSet, err := initPlanFromFile(planFile)
		if err != nil {
			fmt.Println("error initializing plan from file")
			os.Exit(1)
		}
		fmt.Printf("A change plan has been loaded and is shown below. These changes have been saved to %s\n\n",
			defaultPlanFile)
		for _, plan := range changeSet.Plans {
			DisplayChangePlanForDirectory(plan)
		}

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
			}
		} else {
			fmt.Println("No Changes found.")
			DisplayChangeCount(changeSet)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVarP(&planFile, "file", "f", defaultPlanFile, "Target a plan file")

	remoteType = defaultRemoteType
}
