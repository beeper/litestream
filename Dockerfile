FROM golang:1.21.3 as builder

WORKDIR /src/litestream
COPY . .

ARG LITESTREAM_VERSION=latest

RUN --mount=type=cache,target=/root/.cache/go-build \
	--mount=type=cache,target=/go/pkg \
	go build -ldflags "-s -w -X 'main.Version=${LITESTREAM_VERSION}' -extldflags '-static'" -tags osusergo,netgo,sqlite_omit_load_extension -o /usr/local/bin/litestream ./cmd/litestream

# to make copy a single layer later
COPY etc/sqlite3 etc/aws-k8s-sa-provider /usr/local/bin/

FROM alpine:3.17.2
ENV AWS_SDK_LOAD_CONFIG=1 \
    AWS_CONFIG_FILE=/etc/aws-config

# for debugging and cleanup
RUN apk add --no-cache sqlite curl aws-cli && \
    mkdir /root/.aws

COPY etc/aws-config /etc/
COPY --from=builder \
     /usr/local/bin/litestream \
     /usr/local/bin/sqlite3 \
     /usr/local/bin/aws-k8s-sa-provider \
     /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/litestream"]
CMD []
