# gtrack
Jira + Toggle CLI time tracker

### Authorization
For authorization you need API key for Toggl. You can get it on https://track.toggl.com/profile page  
Use `gtrack auth`, and follow prompts (you will need to provide a domain, login, password for Jira, and API key and select workspace for Toggl)  

### List your issues
Use `gtrack list` for viewing your Jira issues with uncommited time  
You can see current tracking issue in this list  

### Timetracking
Use `gtrack start` and select task with Up/Down arrow keys for starting timetracking  
You can have only one "current" issue, just start another
You can stop current timetrack with `gtrack stop`

### Commit your worklogs!
For commiting your worklogs (you like commiting?) use `gtrack commit`
