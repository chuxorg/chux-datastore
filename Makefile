build:
	go build -o chux-mongo ./...
.PHONY: test
test:
	go test ./...

.PHONY: release-version
release-version:
	./scripts/release_version.sh


.PHONY: changelog
changelog:
	chmod +x ./scripts/changelog.sh
	./scripts/changelog.sh

