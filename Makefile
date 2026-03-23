OUT:=$(CURDIR)/out

.PHONY: all build outdir test coverage fmt vet clean ko

test:
	go test ./...

outdir:
	-mkdir -p $(OUT)

coverage: outdir
	go test -coverprofile=$(OUT)/coverage.out ./...
	go tool cover -html="$(OUT)/coverage.out" -o $(OUT)/coverage.html

fmt:
	go fmt ./...

checkfmt:
	@files=$$(gofmt -l ./); \
	if [ -n "$$files" ]; then \
		printf '%s\n' "$$files"; \
		exit 1; \
	fi


vet:
	go vet ./...

clean:
	-rm -rf $(OUT)
