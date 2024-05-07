.PHONY: test
test:
	go test ./...

.PHONY: watch_test
watch_test:
	find . | entr -c -r go test ./...

watch_dev_server:
	find . | entr -c -r go run main.go config.dev.yaml
