#!/usr/bin/osascript

# Stacklane Manager - Standalone Application
# Double-click this file to manage your Stacklane workflow

try
    # Main menu dialog
    set menuChoice to choose from list {"🚀 Start Stack", "🛑 Stop Stack", "📊 View Status", "📋 View Logs", "❌ Cancel"} with title "Stacklane Manager" with prompt "What would you like to do?" default items {"🚀 Start Stack"}
    
    if menuChoice is false or menuChoice = {"❌ Cancel"} then
        return
    end if
    
    set action to item 1 of menuChoice
    
    if action = "🚀 Start Stack" then
        startStack()
    else if action = "🛑 Stop Stack" then
        stopStack()
    else if action = "📊 View Status" then
        viewStatus()
    else if action = "📋 View Logs" then
        viewLogs()
    end if
    
on error errMsg
    display alert "❌ Error" message errMsg buttons {"OK"} default button "OK"
end try

# Function to start the stack
on startStack()
    try
        # Get project directory
        set projectPath to choose folder with prompt "📁 Select your project directory:"
        set projectPath to POSIX path of projectPath
        
        # Get project name for display
        set projectName to basename(projectPath)
        
        # Ask for custom settings
        set settingsDialog to display dialog "⚙️ Custom settings (optional):" default answer "HOST_PORT=80" with title "Stacklane Settings" buttons {"Skip", "Use Settings"} default button "Skip"
        
        set useCustomSettings to button returned of settingsDialog = "Use Settings"
        set customSettings to ""
        if useCustomSettings then
            set customSettings to text returned of settingsDialog
        end if
        
        # Build the command
        set shellScript to "cd '" & projectPath & "';" & return
        
        if customSettings is not "" then
            set shellScript to shellScript & "export " & customSettings & ";" & return
        end if
        
        set shellScript to shellScript & "export COMPOSE_PROJECT_NAME='" & projectName & "';" & return
        set shellScript to shellScript & "export CODE_DIR='" & projectPath & "';" & return
        set shellScript to shellScript & "echo '🚀 Starting Stacklane for project: " & projectName & "';" & return
        set shellScript to shellScript & "echo '📁 Code directory: " & projectPath & "';" & return
        set shellScript to shellScript & "docker compose -f '$HOME/docker/20i-stack/docker-compose.yml' up -d;" & return
        set shellScript to shellScript & "echo '✅ Stack started! Access your site at: http://localhost';" & return
        set shellScript to shellScript & "echo '🔧 phpMyAdmin: http://localhost:8081';"
        
        # Execute in Terminal
        tell application "Terminal"
            activate
            do script shellScript
        end tell
        
        # Show success notification
        display notification "Stack starting for: " & projectName with title "🚀 Stacklane" subtitle "Check Terminal for details"
        
    on error errMsg
        display alert "❌ Error Starting Stack" message errMsg buttons {"OK"} default button "OK"
    end try
end startStack

# Function to stop the stack
on stopStack()
    try
        # Get list of running compose projects
        set shellScript to "docker ps --format 'table {{.Names}}' | grep -E '^[^-]+-[^-]+-[0-9]+$' | sed 's/-[^-]*-[0-9]*$//' | sort -u"
        set runningProjects to do shell script shellScript
        
        if runningProjects = "" then
            display alert "ℹ️ No Running Stacks" message "No Stacklane projects appear to be running." buttons {"OK"} default button "OK"
            return
        end if
        
        # Convert to list for dialog
        set projectList to paragraphs of runningProjects
        set selectedProject to choose from list projectList with title "🛑 Stop Stacklane Project" with prompt "Select project to stop:" default items {item 1 of projectList}
        
        if selectedProject is false then
            return
        end if
        
        set projectName to item 1 of selectedProject
        
        set shellScript to "export COMPOSE_PROJECT_NAME='" & projectName & "';" & return
        set shellScript to shellScript & "echo '🛑 Stopping Stacklane project: " & projectName & "';" & return
        set shellScript to shellScript & "docker compose -f '$HOME/docker/20i-stack/docker-compose.yml' down;" & return
        set shellScript to shellScript & "echo '✅ Stack stopped: " & projectName & "';"
        
        tell application "Terminal"
            activate
            do script shellScript
        end tell
        
        display notification "Stack stopped: " & projectName with title "🛑 Stacklane"
        
    on error errMsg
        display alert "❌ Error Stopping Stack" message errMsg buttons {"OK"} default button "OK"
    end try
end stopStack

# Function to view status
on viewStatus()
    try
        set shellScript to "echo '📊 Stacklane Status:';" & return
        set shellScript to shellScript & "docker compose -f '$HOME/docker/20i-stack/docker-compose.yml' ps;" & return
        set shellScript to shellScript & "echo '';" & return
        set shellScript to shellScript & "echo '🐳 All Docker containers:';" & return
        set shellScript to shellScript & "docker ps --format 'table {{.Names}}\\t{{.Status}}\\t{{.Ports}}' | head -20"
        
        tell application "Terminal"
            activate
            do script shellScript
        end tell
        
    on error errMsg
        display alert "❌ Error Viewing Status" message errMsg buttons {"OK"} default button "OK"
    end try
end viewStatus

# Function to view logs
on viewLogs()
    try
        # Get list of running compose projects
        set shellScript to "docker ps --format 'table {{.Names}}' | grep -E '^[^-]+-[^-]+-[0-9]+$' | sed 's/-[^-]*-[0-9]*$//' | sort -u"
        set runningProjects to do shell script shellScript
        
        if runningProjects = "" then
            display alert "ℹ️ No Running Stacks" message "No Stacklane projects appear to be running." buttons {"OK"} default button "OK"
            return
        end if
        
        # Convert to list for dialog
        set projectList to paragraphs of runningProjects
        set selectedProject to choose from list projectList with title "📋 View Stacklane Logs" with prompt "Select project to view logs:" default items {item 1 of projectList}
        
        if selectedProject is false then
            return
        end if
        
        set projectName to item 1 of selectedProject
        
        set shellScript to "export COMPOSE_PROJECT_NAME='" & projectName & "';" & return
        set shellScript to shellScript & "echo '📋 Viewing logs for: " & projectName & "';" & return
        set shellScript to shellScript & "echo 'Press Ctrl+C to stop following logs';" & return
        set shellScript to shellScript & "docker compose -f '$HOME/docker/20i-stack/docker-compose.yml' logs -f"
        
        tell application "Terminal"
            activate
            do script shellScript
        end tell
        
    on error errMsg
        display alert "❌ Error Viewing Logs" message errMsg buttons {"OK"} default button "OK"
    end try
end viewLogs

# Helper function to get basename
on basename(posixPath)
    set AppleScript's text item delimiters to "/"
    set pathItems to text items of posixPath
    set AppleScript's text item delimiters to ""
    
    # Remove trailing slash if present
    if item -1 of pathItems = "" then
        return item -2 of pathItems
    else
        return item -1 of pathItems
    end if
end basename
