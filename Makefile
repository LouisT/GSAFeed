include .env

LDFLAGS=-s -w -X 'main.Project=${PROJECT}' -X 'main.Version=${VERSION}' -X 'main.Revision=${REV}'

DEVLDFLAGS=-X 'main.Project=${PROJECT}-dev'

run:
	go build -trimpath -o ./bin/GSFeed -v -ldflags "${LDFLAGS}" ; ./bin/GSFeed --config config.hjson

dev:
	go build -trimpath -o ./bin/GSFeed-dev -v -ldflags "${LDFLAGS} ${DEVLDFLAGS}" ; ./bin/GSFeed-dev --config config.dev.hjson