FROM golang:1.12 as build-env

WORKDIR /go/src/app
ADD . /go/src/app

RUN GO111MODULE=on go build -o /go/bin/app

FROM gcr.io/distroless/base
COPY --from=build-env /go/bin/app /
CMD ["/app"]