build:
	go build -o chux-mongo ./...
.PHONY: test
test:
	go test ./...

.PHONY: release-version
release-version:
	./release_version.sh


.PHONY: changelog
changelog:
	echo "# Changelog" > CHANGELOG.md
	git tag --sort=-version:refname | while read -r TAG; do \
	  echo -e "\n## $$TAG\n" >> CHANGELOG.md; \
	  if [ "$$PREVIOUS_TAG" != "" ]; then \
	    git log --no-merges --format="* %s (%h)" $$TAG..$$PREVIOUS_TAG >> CHANGELOG.md; \
	  else \
	    git log --no-merges --format="* %s (%h)" $$TAG >> CHANGELOG.md; \
	  fi; \
	  PREVIOUS_TAG=$$TAG; \
	done