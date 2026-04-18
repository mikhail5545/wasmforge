.PHONY: build-ui build-go clean bench-build bench-run

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

bench-build:
	go build -o bin/wasmforge cmd/gateway/main.go
	go build -o bench/bin/upstream ./bench/cmd/upstream
	go build -o bench/bin/native-gateway ./bench/cmd/native-gateway

bench-run: bench-build
	./bench/scripts/run_suite.sh
