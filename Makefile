.PHONY: all test clean zip mac

### バージョンの定義
VERSION     := "v1.0.0"
COMMIT      := $(shell git rev-parse --short HEAD)

### コマンドの定義
GO          = go
GO_BUILD    = $(GO) build
GO_TEST     = $(GO) test -v
GO_LDFLAGS  = -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)"
ZIP          = zip

### ターゲットパラメータ
DIST = dist
SRC = ./main.go 
TARGETS     = $(DIST)/twhr2trap.exe $(DIST)/twhr2trap.app $(DIST)/twhr2trap $(DIST)/twhr2trap.arm
GO_PKGROOT  = ./...

### PHONY ターゲットのビルドルール
all: $(TARGETS)
test:
	env GOOS=$(GOOS) $(GO_TEST) $(GO_PKGROOT)
clean:
	rm -rf $(TARGETS) $(DIST)/*.zip
mac: $(DIST)/twhr2trap.app
zip: $(TARGETS)
	cd dist && $(ZIP) twhr2trap_win.zip twhr2trap.exe
	cd dist && $(ZIP) twhr2trap_mac.zip twhr2trap.app
	cd dist && $(ZIP) twhr2trap_linux_amd64.zip twhr2trap
	cd dist && $(ZIP) twhr2trap_linux_arm.zip twhr2trap.arm

### 実行ファイルのビルドルール
$(DIST)/twhr2trap.exe: $(SRC)
	env GO111MODULE=on GOOS=windows GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twhr2trap.app: $(SRC)
	env GO111MODULE=on GOOS=darwin GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twhr2trap.arm: $(SRC)
	env GO111MODULE=on GOOS=linux GOARCH=arm GOARM=7 $(GO_BUILD) $(GO_LDFLAGS) -o $@
$(DIST)/twhr2trap: $(SRC)
	env GO111MODULE=on GOOS=linux GOARCH=amd64 $(GO_BUILD) $(GO_LDFLAGS) -o $@

hash:
	cd dist && shasum -a 256 *
