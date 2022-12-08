include .env

LDFLAGS=-s -w -X 'main.Project=${PROJECT}' -X 'main.Version=${VERSION}' -X 'main.Revision=${REV}'

DEVLDFLAGS=-X 'main.Project=${PROJECT}-dev'

run:
	go build -trimpath -o ./bin/GSAFeed -v -ldflags "${LDFLAGS}" ; ./bin/GSAFeed --config config.hjson

dev:
	go build -trimpath -o ./bin/GSAFeed-dev -v -ldflags "${LDFLAGS} ${DEVLDFLAGS}" ; ./bin/GSAFeed-dev --config config.dev.hjson