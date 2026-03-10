OUT:=$(CURDIR)/out

.PHONY: all build outdir test coverage fmt vet clean ko

outdir:
	-mkdir -p $(OUT)

test:
	go test ./...

coverage: outdir
	go test -coverprofile=$(OUT)/coverage.out ./...
	go tool cover -html="$(OUT)/coverage.out" -o $(OUT)/coverage.html

fmt:
	go fmt ./...

checkfmt:
	@gofmt -l ./


vet:
	go vet ./...

clean:
	-rm -rf $(OUT)
