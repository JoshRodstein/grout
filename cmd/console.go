package cmd

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
)

const (
	// grout ASCII Banner and Sub title
	GROUT      = "\n  _____ _____   ____  _    _ _______\n / ____|  __ \\ / __ \\| |  | |__   __|\n| |  __| |__) | |  | | |  | |  | |\n| | |_ |  _  /| |  | | |  | |  | |\n| |__| | | \\ \\| |__| | |__| |  | | \n \\_____|_|  \\_\\\\____/ \\____/   |_|  \n\n"
	SpellItOUt = "Github Remote Migration Utility"
)

func IAmGrout() {
	fmt.Println(GROUT + SpellItOUt)
}

func DisplayBundledErrors() {
	var output string
	for i := 0; i < errorBundle.Count; i++ {
		errorOut := fmt.Sprintf("\n\terror %d: %s\n", i+1, errorBundle.Errors[i])
		output += errorOut
	}
	fmt.Print(output)
}

func DisplayBundledErrorsPlan() {
	if errorBundle.Count > 0 {
		fmt.Println(": grout encountered the following error(s) while generating a plan...")
		DisplayBundledErrors()
		fmt.Println("You should review these errors before executing an update, as they may effect your desired changes.")
		fmt.Println("----------------------------")
	}
}

func DisplayBundledErrorsUpdate() {
	if errorBundle.Count > 0 {
		fmt.Println("----> grout encountered the following error(s) while executing an update...")
		DisplayBundledErrors()
		fmt.Println("\nYou should review these errors to ensure that your desired changes have been made.")
		fmt.Println("----------------------------")
	}
}

func DisplayChangeCount(changes ChangeSet) {
	fmt.Printf("\nPlanned %d change(s) across %d repo(s)\n\n", changes.Count, len(changes.Plans))
}

func DisplayChangeIntention(changes ChangeSet) {
	fmt.Printf("\ngrout will perform %d change(s) across %d repo(s)\n\n", changes.Count, len(changes.Plans))
}

func DisplayChangeResult(changes ChangeSet) {
	fmt.Printf("\nCompleted %d change(s) across %d repo(s)\n\n", changes.Count, len(changes.Plans))
}

func DisplayChangePlanForDirectory(plan RepoPlan) {
	localRepo := plan.Repo

	var sb strings.Builder
	output := fmt.Sprintf(""+
		"%sRepository:   %s\n"+
		"%sPath:\t\t%s",
		twoSpaces, localRepo.Name,
		twoSpaces, localRepo.Path)
	sb.WriteString(output)
	fmt.Println(sb.String())
	for _, change := range plan.Changes {
		fmt.Printf("%sRemote: \t%s\n", sixSpaces, change.Name)
		for i := 0; i < len(change.NewURLs); i++ {
			fmt.Printf("%s  Change:       %s -> %s\n", sixSpaces, change.CurrentURLs[i], change.NewURLs[i])
		}
	}
	fmt.Println()
}

func ParametersConfirmationOutput() string {
	targetOrgVal := targetOrganization
	if len(targetOrganization) == 0 {
		targetOrgVal = "all orgs"
	}
	newOrgVal := newOrganization
	if len(newOrganization) == 0 {
		newOrgVal = "unchanged"
	}
	confirmation := fmt.Sprintf(
		"\n    Search Directory:      %s\n"+
			"    Target URL:            %s\n"+
			"    New URL:               %s\n"+
			"    Target Organization:   %s\n"+
			"    New Organization:      %s\n"+
			"\nEnter '%s' to confirm parameters and create a plan: ",
		targetDir, targetRemoteURL, newRemoteURL, targetOrgVal, newOrgVal, Yes)
	return confirmation
}

func promptForInput(prompt string, dflt string) string {
	var reply string
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)

	if runtime.GOOS == Windows {
		reply, _ = reader.ReadString('\r')
		reply = strings.Replace(reply, "\r", "", -1)
	} else {
		reply, _ = reader.ReadString('\n')
		reply = strings.Replace(reply, "\n", "", -1)
	}

	if len(reply) == 0 {
		reply = dflt
	}

	return reply
}
