build-all:
	@echo "Update CLI for WSL2..."
	go build -o focus ./cmd/cli
	sudo cp focus /usr/local/bin/
	
	@echo "Update CLI & Blocker for Windows..."
	GOOS=windows go build -o focus.exe ./cmd/cli
	GOOS=windows go build -o blocker.exe ./cmd/blocker

	
	@echo "Successfully"