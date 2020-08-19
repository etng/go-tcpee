snapshot:
	goreleaser build --snapshot --rm-dist
build:
	goreleaser build --rm-dist
release:
	goreleaser release --rm-dist
dev:
	air
