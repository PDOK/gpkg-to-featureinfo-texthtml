FROM golang:1.13-alpine3.10

RUN apk update && apk upgrade && \
    apk add --no-cache bash git gcc musl-dev

ENV GO113MODULE=on

WORKDIR /go/src/service

ADD ./ /go/src/service

#disable crosscompiling
ENV CGO_ENABLED=1

#compile linux only
ENV GOOS=linux

#build the binary with debug information removed
RUN go build -ldflags '-w -s -linkmode external -extldflags -static' -a -installsuffix cgo -o /gpkg-to-featureinfo-texthtml gpkg-to-featureinfo-texthtml

FROM scratch
WORKDIR /
ENV PATH=/

COPY --from=0 /gpkg-to-featureinfo-texthtml /

CMD ["gpkg-to-featureinfo-texthtml"]