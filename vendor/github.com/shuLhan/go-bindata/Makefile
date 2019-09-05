##
## This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain
## Dedication license. Its contents can be found at:
## http://creativecommons.org/publicdomain/zero/1.0
##

.PHONY: all build
.PHONY: test coverbrowse test-cmd
.PHONY: clean distclean

CMD_DIR           :=./cmd/go-bindata
TESTDATA_DIR      :=./testdata
TESTDATA_IN_DIR   :=./testdata/in
TESTDATA_OUT_DIR  :=./testdata/out

SRC               :=$(shell go list -f '{{ $$dir := .Dir }}{{ range .GoFiles }} {{ $$dir }}/{{.}} {{end}}' ./...)
TEST              :=$(shell go list -f '{{ $$dir := .Dir }}{{ range .TestGoFiles }} {{ $$dir }}/{{.}} {{end}}' ./...)
TEST_COVER_ALL    :=cover.out

TARGET_CMD        :=$(shell go list -f '{{ .Target }}' $(CMD_DIR))
CMD_SRC           :=$(shell go list -f '{{ $$dir := .Dir }}{{ range .GoFiles }} {{ $$dir }}/{{.}} {{end}}' $(CMD_DIR))
CMD_TEST          :=$(shell go list -f '{{ $$dir := .Dir }}{{ range .TestGoFiles }} {{ $$dir }}/{{.}} {{end}}' $(CMD_DIR))
CMD_COVER_OUT     :=$(CMD_DIR)/cover.out

TARGET_LIB        :=$(shell go list -f '{{ .Target }}' ./)
LIB_SRC           :=$(shell go list -f '{{ $$dir := .Dir }}{{ range .GoFiles }} {{ $$dir }}/{{.}} {{end}}' ./)
LIB_TEST          :=$(shell go list -f '{{ $$dir := .Dir }}{{ range .TestGoFiles }} {{ $$dir }}/{{.}} {{end}}' ./)
LIB_COVER_OUT     :=lib.cover.out

TEST_COVER_OUT    :=$(TEST_COVER_ALL) $(LIB_COVER_OUT) $(CMD_COVER_OUT)
TEST_COVER_HTML   :=cover.html

POST_TEST_FILES   := \
	./assert_test.go \
	$(TESTDATA_DIR)/_bindata_test.go \
	$(TESTDATA_DIR)/_out_default_single.go

TEST_OUT          := \
	$(TESTDATA_OUT_DIR)/opt/no-output/bindata.go \
	$(TESTDATA_OUT_DIR)/default/single/bindata.go \
	$(TESTDATA_OUT_DIR)/compress/memcopy/bindata.go \
	$(TESTDATA_OUT_DIR)/compress/nomemcopy/bindata.go \
	$(TESTDATA_OUT_DIR)/debug/bindata.go \
	$(TESTDATA_OUT_DIR)/nocompress/memcopy/bindata.go \
	$(TESTDATA_OUT_DIR)/nocompress/nomemcopy/bindata.go \
	$(TESTDATA_OUT_DIR)/split/bindata.go \
	$(TESTDATA_OUT_DIR)/symlinkFile/bindata.go \
	$(TESTDATA_OUT_DIR)/symlinkParent/bindata.go \
	$(TESTDATA_OUT_DIR)/symlinkRecursiveParent/bindata.go

VENDOR_DIR        :=$(PWD)/vendor
VENDOR_BIN        :=$(VENDOR_DIR)/bin


##
## MAIN TARGET
##

all: build test-cmd

##
## CLEAN
##

clean:
	rm -rf $(TEST_COVER_OUT) $(TEST_COVER_HTML) $(TESTDATA_OUT_DIR)

distclean: clean
	rm -rf $(TARGET_CMD) $(TARGET_LIB) $(VENDOR_DIR)


##
## TEST
##

$(LIB_COVER_OUT): $(LIB_SRC) $(LIB_TEST)
	@echo ""
	@echo ">>> Testing library ..."
	@go test -v ./ && \
		go test -coverprofile=$@ ./ &>/dev/null && \
		go tool cover -func=$@


$(CMD_COVER_OUT): $(CMD_SRC) $(CMD_TEST)
	@echo ""
	@echo ">>> Testing cmd ..."
	@go test -v $(CMD_DIR) && \
		go test -coverprofile=$@ $(CMD_DIR) &>/dev/null && \
		go tool cover -func=$@

$(TEST_COVER_ALL): $(LIB_COVER_OUT) $(CMD_COVER_OUT)
	@echo ""
	@echo ">>> Generate single coverage '$@' ..."
	@cat $^ | sed '/mode: set/d' | sed '1s/^/mode: set\n/' > $@

$(TEST_COVER_HTML): $(TEST_COVER_ALL)
	@echo ">>> Generate HTML coverage '$@' ..."
	@go tool cover -html=$< -o $@

test: $(TEST_COVER_HTML)

coverbrowse: test
	@xdg-open $(TEST_COVER_HTML)


##
## BUILD
##

$(TARGET_LIB): $(LIB_SRC)
	go install ./

$(TARGET_CMD): $(CMD_SRC) $(TARGET_LIB)
	go install $(CMD_DIR)

build: test $(TARGET_CMD)


##
## TEST POST BUILD
##

$(TESTDATA_OUT_DIR)/opt/no-output/bindata.go: OUT_DIR=$(TESTDATA_OUT_DIR)/opt/no-output
$(TESTDATA_OUT_DIR)/opt/no-output/bindata.go: $(TESTDATA_IN_DIR)/*
	@echo ""
	@echo ">>> Testing without '-o' flag"
	mkdir -p $(OUT_DIR)
	cd $(OUT_DIR) \
		&& $(TARGET_CMD) -pkg bindata -prefix=".*/testdata/" \
			-ignore="split/" ../../../../$(TESTDATA_IN_DIR)/...
	cp ./assert_test.go $(OUT_DIR)
	cp $(TESTDATA_DIR)/_bindata_test.go $(OUT_DIR)/bindata_test.go
	go test -v $(OUT_DIR)

$(TESTDATA_OUT_DIR)/compress/memcopy/bindata.go: OUT_DIR=$(TESTDATA_OUT_DIR)/compress/memcopy
$(TESTDATA_OUT_DIR)/compress/memcopy/bindata.go: $(TESTDATA_IN_DIR)/*
	@echo ""
	@echo ">>> Testing default option (compress, memcopy)"
	mkdir -p $(OUT_DIR)
	$(TARGET_CMD) -o $@ -pkg bindata -prefix=".*testdata/" \
		-ignore="split/" $(TESTDATA_IN_DIR)/...
	cp ./assert_test.go $(OUT_DIR)
	cp $(TESTDATA_DIR)/_bindata_test.go $(OUT_DIR)/bindata_test.go
	go test -v $(OUT_DIR)

$(TESTDATA_OUT_DIR)/default/single/bindata.go: OUT_DIR=$(TESTDATA_OUT_DIR)/default/single
$(TESTDATA_OUT_DIR)/default/single/bindata.go: $(TESTDATA_IN_DIR)/*
	@echo ""
	@echo ">>> Testing default option with single input file"
	mkdir -p $(OUT_DIR)
	$(TARGET_CMD) -o $@ -pkg bindata -prefix=".*testdata/" \
		-ignore="split/" $(TESTDATA_IN_DIR)/test.asset
	cp ./assert_test.go $(OUT_DIR)
	cp $(TESTDATA_DIR)/_out_default_single.go $(OUT_DIR)/bindata_test.go
	go test -v $(OUT_DIR) || rm -f $(OUT_DIR)/*

$(TESTDATA_OUT_DIR)/compress/nomemcopy/bindata.go: OUT_DIR=$(TESTDATA_OUT_DIR)/compress/nomemcopy
$(TESTDATA_OUT_DIR)/compress/nomemcopy/bindata.go: $(TESTDATA_IN_DIR)/*
	@echo ""
	@echo ">>> Testing with '-nomemcopy'"
	mkdir -p $(OUT_DIR)
	$(TARGET_CMD) -o $@ -pkg bindata -prefix=".*testdata/" \
		-ignore="split/" -nomemcopy $(TESTDATA_IN_DIR)/...
	cp ./assert_test.go $(OUT_DIR)
	cp $(TESTDATA_DIR)/_bindata_test.go $(OUT_DIR)/bindata_test.go
	go test -v $(OUT_DIR)

$(TESTDATA_OUT_DIR)/debug/bindata.go: OUT_DIR=$(TESTDATA_OUT_DIR)/debug
$(TESTDATA_OUT_DIR)/debug/bindata.go: $(TESTDATA_IN_DIR)/*
	@echo ""
	@echo ">>> Testing opt 'debug'"
	mkdir -p $(OUT_DIR)
	$(TARGET_CMD) -o $@ -pkg bindata -prefix=".*testdata/" \
		-ignore="split/" -debug $(TESTDATA_IN_DIR)/...
	cp ./assert_test.go $(OUT_DIR)
	cp $(TESTDATA_DIR)/_bindata_test.go $(OUT_DIR)/bindata_test.go
	go test -v $(OUT_DIR)

$(TESTDATA_OUT_DIR)/nocompress/memcopy/bindata.go: OUT_DIR=$(TESTDATA_OUT_DIR)/nocompress/memcopy
$(TESTDATA_OUT_DIR)/nocompress/memcopy/bindata.go: $(TESTDATA_IN_DIR)/*
	@echo ""
	@echo ">>> Testing opt '-nocompress'"
	mkdir -p $(OUT_DIR)
	$(TARGET_CMD) -o $@ -pkg bindata -prefix=".*testdata/" \
		-ignore="split/" -nocompress $(TESTDATA_IN_DIR)/...
	cp ./assert_test.go $(OUT_DIR)
	cp $(TESTDATA_DIR)/_bindata_test.go $(OUT_DIR)/bindata_test.go
	go test -v $(OUT_DIR)

$(TESTDATA_OUT_DIR)/nocompress/nomemcopy/bindata.go: OUT_DIR=$(TESTDATA_OUT_DIR)/nocompress/nomemcopy
$(TESTDATA_OUT_DIR)/nocompress/nomemcopy/bindata.go: $(TESTDATA_IN_DIR)/*
	@echo ""
	@echo ">>> Testing opt '-nocompress -nomemcopy'"
	mkdir -p $(OUT_DIR)
	$(TARGET_CMD) -o $@ -pkg bindata -prefix=".*testdata/" \
		-ignore="split/" -nocompress -nomemcopy $(TESTDATA_IN_DIR)/...
	cp ./assert_test.go $(OUT_DIR)
	cp $(TESTDATA_DIR)/_bindata_test.go $(OUT_DIR)/bindata_test.go
	go test -v $(OUT_DIR)

$(TESTDATA_OUT_DIR)/split/bindata.go: OUT_DIR=$(TESTDATA_OUT_DIR)/split
$(TESTDATA_OUT_DIR)/split/bindata.go: $(TESTDATA_DIR)/_split_test.go $(TESTDATA_IN_DIR)/split/*
	@echo ""
	@echo ">>> Testing opt '-split'"
	mkdir -p $(OUT_DIR)
	rm -f $(OUT_DIR)/*
	$(TARGET_CMD) -o $(OUT_DIR) -pkg bindata -prefix=".*testdata/" \
		-split $(TESTDATA_IN_DIR)/split/...
	cp ./assert_test.go $(OUT_DIR)
	cp $< $(OUT_DIR)/bindata_test.go
	go test -v $(OUT_DIR)

$(TESTDATA_OUT_DIR)/symlinkFile/bindata.go: OUT_DIR=$(TESTDATA_OUT_DIR)/symlinkFile
$(TESTDATA_OUT_DIR)/symlinkFile/bindata.go: $(TESTDATA_DIR)/symlinkFile $(TESTDATA_DIR)/_symlinkFile_test.go
	@echo ""
	@echo ">>> Testing symlink to file"
	mkdir -p $(OUT_DIR)
	rm -f $(OUT_DIR)/*
	$(TARGET_CMD) -o $@ -pkg bindata -prefix=".*testdata/" \
		$(TESTDATA_DIR)/symlinkFile/...
	cp ./assert_test.go $(OUT_DIR)
	cp $(TESTDATA_DIR)/_symlinkFile_test.go $(OUT_DIR)/bindata_test.go
	go test -v $(OUT_DIR)

$(TESTDATA_OUT_DIR)/symlinkParent/bindata.go: OUT_DIR=$(TESTDATA_OUT_DIR)/symlinkParent
$(TESTDATA_OUT_DIR)/symlinkParent/bindata.go: $(TESTDATA_DIR)/_symlinkParent_test.go $(TESTDATA_DIR)/symlinkParent
	@echo ""
	@echo ">>> Testing symlink to directory"
	mkdir -p $(OUT_DIR)
	rm -f $(OUT_DIR)/*
	$(TARGET_CMD) -o $@ -pkg bindata -prefix=".*testdata/" \
		$(TESTDATA_DIR)/symlinkParent/...
	cp ./assert_test.go $(OUT_DIR)
	cp $< $(OUT_DIR)/bindata_test.go
	go test -v $(OUT_DIR)

$(TESTDATA_OUT_DIR)/symlinkRecursiveParent/bindata.go: OUT_DIR=$(TESTDATA_OUT_DIR)/symlinkRecursiveParent
$(TESTDATA_OUT_DIR)/symlinkRecursiveParent/bindata.go: $(TESTDATA_DIR)/_symlinkRecursiveParent_test.go $(TESTDATA_DIR)/symlinkRecursiveParent
	@echo ""
	@echo ">>> Testing symlink recursive directory"
	mkdir -p $(OUT_DIR)
	rm -f $(OUT_DIR)/*
	$(TARGET_CMD) -o $@ -pkg bindata -prefix=".*testdata/" \
		$(TESTDATA_DIR)/symlinkRecursiveParent/...
	cp ./assert_test.go $(OUT_DIR)
	cp $< $(OUT_DIR)/bindata_test.go
	go test -v $(OUT_DIR)

$(TEST_OUT): $(TARGET_CMD) $(POST_TEST_FILES)

test-cmd: $(TEST_OUT)
