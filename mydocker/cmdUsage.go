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
)

const (
	ttyUsage         = "Allocate a pseudo-TTY"
	interactiveUsage = "Keep STDIN open even if not attached"
	memoryUsage      = "Memory limit"
)
