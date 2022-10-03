# GROUT
Git Remote Migration Utility

Welcome to the Git Remote Migration Utility. This command line utility is designed to make finding and changing git remotes easier and less confusing.

## What is GROUT?
GROUT is a **command line utility** for updating local remotes across one or many local repos.

## What can GROUT do?   

   - Target remotes under a specific org.
   - Change the organization for targeted remotes.
   - Target a specific remote URL.
   - Produce an easily readable JSON plan of the proposed changes for review before updating.

## How do I use GROUT?

Locate the binary for your current operating system and run in one of the following two modes... 
 
### Interactive (Recommended)

Run the 'grout-<my_OS>' command and follow the prompts to specify...


  ```
  $ grout-macos

  Running interactive mode.

   _____ _____   ____  _    _ _______ 
  / ____|  __ \ / __ \| |  | |__   __|
 | |  __| |__) | |  | | |  | |  | |   
 | | |_ |  _  /| |  | | |  | |  | |   
 | |__| | | \ \| |__| | |__| |  | |   
  \_____|_|  \_\\____/ \____/   |_|    

  Github Remote Migration Utility
  ---------------------
  Current Remote(s) URL (github.com): 
  New Remote(s) URL (github.com):
  Target Github org/owner (All orgs if not set):
  New Github org/owner (unchanged if not set):
  Search directory (default to current directory):
  ```

Grout will then display your selections and prompt for confirmation before creating a migration plan...

```
    Search Directory:      /Users/username/my/current/directory/{repo}
    Target URL:            github.com
    New URL:               github.com
    Target Org/Owner:      all orgs/owners
    New Org/Owner:         unchanged


Enter 'Yes' to confirm selections and create a plan:
```

Grout will then generate and display a migration plan for your Git remotes...

```
Enter 'Yes' to confirm: Yes

Generating plan...
A change plan has been generated and is shown below. These changes have been saved to grout-plan.json

  Repository:   grout
  Path:         /Users/username/my/current/directory/{repo}/.git
  Remote:       origin
    Change:       http://github.com/{owner}/{repo}.git -> https://github.com/{owner}/{repo}.git


GROUT will perform 2 change(s) across 1 repo(s)

---------------------
Enter 'y' to accept and apply these changes: 
 'Yes' to accept and apply these changes:
```

Once the proposed changes have been reviewed and accepted, Grout will execute and confirm the plan and its results

```
Completed 1 change(s) across 1 repo(s)
```

You can re-confirm the changes by running interactive mode again. The plan should result in...

```
No changes found
```

### Non-interactive

#### Plan
    Generate a plan for updating git remotes:
    
      Search all folders under given directory for git repositories and 
      Spit out a plan for updating git remotes to a new URL. Plan is saved 
      as test-grout-plan.json unless otherwise specified.
    
    Usage:
      grout plan [flags]
    
    Flags:
      -d, --directory string   Set search directory
          --find-org string    set target org for remote update
          --find-url string    set remote url for remote update (default "github.com")
      -h, --help               help for plan
          --set-org string     set target org for remote update
          --set-url string     set target url for remote update (default "github.com")
      -t, --toggle             Help message for toggle
    
    Global Flags:
          --config string   config file (default is $HOME/.grut_bin.yaml)
      -v, --verbose         Verbose output for logging/debugging

      
#### Update
    Execute changes in a plan:
            
      Load and execute a plan from file. GROUT looks for test-grout-plan.json in 
      it's current directory unless otherwise specified.
    
    Usage:
      grout update [flags]
    
    Flags:
      -f, --file string   Target a plan file (default "grout-plan.json")
      -h, --help          help for update
    
    Global Flags:
          --config string   config file (default is $HOME/.grut_bin.yaml)
      -v, --verbose         Verbose output for logging/debugging





