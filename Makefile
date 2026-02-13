.PHONY: build-ui build-go clean

build-ui:
	cd ui/admin-ui && npm install && npm run build

prepare-embed: build-ui
	# Remove old build if exists
	rm -rf pkg/ui/out
	# Copy the fresh build to where Go expects it
	cp -r ui/admin-ui/out pkg/ui/out

build: prepare-embed
	go build -o gateway cmd/gateway/main.go

clean:
	rm -rf pkg/ui/out
	rm -rf ui/admin-ui/out
	rm -f gateway