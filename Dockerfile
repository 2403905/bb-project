FROM golang:1.19-alpine as scratch

RUN apk add build-base
ENV CGO_ENABLED=1
WORKDIR /src
COPY go.mod go.sum .
RUN go mod download && go mod verify
COPY . .
WORKDIR cmd
RUN go build -tags musl -v -o /go/bin/bb-project


FROM alpine:3.16

WORKDIR /service/cmd
COPY --from=scratch /go/bin/bb-project .

CMD ["./bb-project"]
