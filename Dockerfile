FROM golang:1.19 as builder

WORKDIR /src/litestream
COPY . .

ARG LITESTREAM_VERSION=latest

RUN --mount=type=cache,target=/root/.cache/go-build \
	--mount=type=cache,target=/go/pkg \
	go build -ldflags "-s -w -X 'main.Version=${LITESTREAM_VERSION}' -extldflags '-static'" -tags osusergo,netgo,sqlite_omit_load_extension -o /usr/local/bin/litestream ./cmd/litestream


FROM alpine
RUN apk add --no-cache sqlite # for debugging
COPY --from=builder /usr/local/bin/litestream /usr/local/bin/litestream
ENTRYPOINT ["/usr/local/bin/litestream"]
CMD []
