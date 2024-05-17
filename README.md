# Directory Synchronization Tool
This tool synchronizes files and directories between a source directory and a destination directory. It ensures that the destination directory mirrors the source directory by copying new or modified files from the source to the destination and deleting files in the destination that no longer exist in the source.

# Installation
1. Make sure you have Go installed on your system.
2. Clone the repository:
 https://gitlab.rebrainme.com/golang_users_repos/5732/finaltask
3. Build the project:
go build -o synctool

# Usage
Run the tool with the source and destination directories as arguments:
`./synctool /path/to/srcDir /path/to/dstDir`
