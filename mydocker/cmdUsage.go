package main

const (
	appShort = "mydocker a simple container runtime implementation"
	appLong  = `mydocker a simple container runtime implementation
				The purpose of this project is to learn how docker works and how to write a docker by ourselves
				Enjoy it , just for fun.`

	runCommandShort = "Create a container"
	runCommandLong  = `Create a container whit namespace and cgroups limit 
					mydocker run -ti [Command]`

	initCommandShort = "Init container process"
	initCommandLong  = `Init container process run user's process in container. 
						Do not call it outside`

	psCommandShort = "List containers"
	psCommandLong  = " List containers"

	logsCommandShort = "Fetch the logs of a container"
	logsCommandLong  = "Fetch the logs of a container"

	execCommandShort = "Execute a command in a running container"
	execCommandLong  = "Execute a command in a running container"

	stopCommandShort = "Stop one or more running containers"
	stopCommandLong  = "Stop one or more running containers"

	rmCommandShort = "Remove one or more containers"
	rmCommandLong  = "Remove one or more containers"
)

const (
	ttyUsage         = "Allocate a pseudo-TTY"
	interactiveUsage = "Keep STDIN open even if not attached"
	memoryUsage      = "Memory limit"
	volumeUsage      = " Bind mount a volume"
	nameUsage        = " Assign a name to the container"
	detachUsage      = " Run container in background and print container ID"
)
