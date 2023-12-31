FROM golang:1-alpine3.18 as build

RUN apk update && \
    apk upgrade && \
    apk add make && \
    apk add protoc && \
    apk add gcc && \
    apk add musl-dev && \
    apk add libpcap-dev
WORKDIR /src
COPY . .
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
RUN export PATH="$PATH:$(go env GOPATH)/bin"
RUN go mod tidy

RUN make build_windows

FROM scratch as sysstatssrv

COPY --from=build /src/sysstatssvc /
COPY --from=build /src/config.json /

ENTRYPOINT ["/sysstatssvc"]
