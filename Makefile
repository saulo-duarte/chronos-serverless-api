build-ChronosFunction:
	@echo "Building Go application..."
	CGO_ENABLED=0 go build -o bootstrap main.go