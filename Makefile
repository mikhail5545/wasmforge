.PHONY: clean build

ifeq($(OS),Windows_NT)
BUILD_SCRIPT := powershell -ExecutionPolicy Bypass -File scripts/build.ps1
CLEAN_UI := if exist pkg\ui\out rmdir /s /q pkg\ui\out
CLEAN_UI2 := if exist ui\adminv2\apps\web\out rmdir /s /q ui\adminv2\apps\web\out
CLEAN_BIN := if exist bin\wasmforge.exe del /q bin\wasmforge.exe
else
BUILD_SCRIPT := bash scripts/build.sh
CLEAN_UI := rm -rf pkg/ui/out
CLEAN_UI2 := rm -rf ui/adminv2/apps/web/out
CLEAN_BIN := rm -f bin/wasmforge
endif

clean:
	$(CLEAN_UI)
	$(CLEAN_UI2)
	$(CLEAN_BIN)

build: clean
	$(BUILD_SCRIPT)

