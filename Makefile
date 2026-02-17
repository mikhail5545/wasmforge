.PHONY: build-ui build-go clean

# Just builds the UI, creating the output in ui/admin-ui/out (static files)
build-ui:
	cd ui/admin-ui && npm install && npm run build

# Places the UI build in the correct location for embedding, then builds the Go binary
prepare-embed: build-ui
	# Remove old build if exists
	rm -rf pkg/ui/out
	# Copy the fresh build to where Go expects it
	cp -r ui/admin-ui/out pkg/ui/out

# Builds the Go binary, ensuring the UI is prepared first
build: prepare-embed
	go build -o bin/wasmforge cmd/gateway/main.go

clean:
	rm -rf pkg/ui/out
	rm -rf ui/admin-ui/out
	rm -f wasmforge

# Just builds the Go binary, assuming the UI is already prepared or not needed
build-go:
	go build -o bin/wasmforge cmd/gateway/main.go

# Runs the npm build for the UI without affecting the Go build
npm-run:
	cd ui/admin-ui && npm install && npm run build

# Runs the UI build and Go build in parallel, ensuring the UI is ready before the Go build starts. Default ports are :3000 for the UI and :8080 for the Go server.
run-separate:
	# Run the UI build in the background
	cd ui/admin-ui && npm install && npm run build &
	# Wait for the UI build to finish before building Go
	wait
	go build -o bin/wasmforge cmd/gateway/main.go