#!/usr/bin/osascript

# Stacklane Launcher - macOS Services Menu Automation
# This AppleScript creates dialogs to manage your Stacklane workflow

on run {input, parameters}
    try
        # Main menu dialog
        set menuChoice to choose from list {"Start Stack", "Stop Stack", "View Status", "View Logs", "Cancel"} with title "Stacklane Manager" with prompt "Choose an action:" default items {"Start Stack"}
        
        if menuChoice is false or menuChoice = {"Cancel"} then
            return
        end if
        
        set action to item 1 of menuChoice
        
        if action = "Start Stack" then
            startStack()
        else if action = "Stop Stack" then
            stopStack()
        else if action = "View Status" then
            viewStatus()
        else if action = "View Logs" then
            viewLogs()
        end if
        
    on error errMsg
        display alert "Error" message errMsg
    end try
end run

# Function to start the stack
on startStack()
    # Get project directory
    set projectPath to choose folder with prompt "Select your project directory:"
    set projectPath to POSIX path of projectPath
    
    # Get optional settings
    set customSettings to display dialog "Optional: Custom settings (leave blank for defaults)" default answer "HOST_PORT=80" with title "Stacklane Settings"
    set settings to text returned of customSettings
    
    # Build the command
    set shellScript to "cd '" & projectPath & "' && "
    
    if settings is not "" then
        set shellScript to shellScript & "export " & settings & " && "
    end if
    
    set shellScript to shellScript & "export COMPOSE_PROJECT_NAME=\"$(basename '" & projectPath & "')\" && "
    set shellScript to shellScript & "export CODE_DIR='" & projectPath & "' && "
    set shellScript to shellScript & "docker compose -f '$HOME/docker/20i-stack/docker-compose.yml' up -d"
    
    # Execute in Terminal
    tell application "Terminal"
        activate
        do script shellScript
    end tell
    
    # Show success message
    display notification "Starting Stacklane for: " & (basename(projectPath)) with title "Stacklane"
end startStack

# Function to stop the stack
on stopStack()
    # Get project name
    set projectName to text returned of (display dialog "Enter project name to stop:" default answer "" with title "Stop Stacklane Project")
    
    if projectName = "" then
        display alert "Error" message "Project name is required"
        return
    end if
    
    set shellScript to "export COMPOSE_PROJECT_NAME='" & projectName & "' && docker compose -f '$HOME/docker/20i-stack/docker-compose.yml' down"
    
    tell application "Terminal"
        activate
        do script shellScript
    end tell
    
    display notification "Stopping Stacklane project: " & projectName with title "Stacklane"
end stopStack

# Function to view status
on viewStatus()
    set shellScript to "docker compose -f '$HOME/docker/20i-stack/docker-compose.yml' ps"
    
    tell application "Terminal"
        activate
        do script shellScript
    end tell
end viewStatus

# Function to view logs
on viewLogs()
    # Get project name
    set projectName to text returned of (display dialog "Enter project name for logs:" default answer "" with title "View Stacklane Logs")
    
    if projectName = "" then
        display alert "Error" message "Project name is required"
        return
    end if
    
    set shellScript to "export COMPOSE_PROJECT_NAME='" & projectName & "' && docker compose -f '$HOME/docker/20i-stack/docker-compose.yml' logs -f"
    
    tell application "Terminal"
        activate
        do script shellScript
    end tell
end viewLogs

# Helper function to get basename
on basename(posixPath)
    set AppleScript's text item delimiters to "/"
    set pathItems to text items of posixPath
    set AppleScript's text item delimiters to ""
    
    if item -1 of pathItems = "" then
        return item -2 of pathItems
    else
        return item -1 of pathItems
    end if
end basename
