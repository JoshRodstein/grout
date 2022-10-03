-include .env

PROJECTNAME := grout

compile:
	# 64-Bit Systems
	# MacOS
	GOOS=darwin GOARCH=amd64 go build -o ./$(PROJECTNAME)-macos main.go
	# Linux
	GOOS=linux GOARCH=amd64 go build -o ./$(PROJECTNAME)-linux main.go
	# Windows
	GOOS=windows GOARCH=amd64 go build -o ./$(PROJECTNAME)-windows.exe main.go