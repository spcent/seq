FROM golang:1.12 as builder
WORKDIR /gowork/github.com/spcent/seq
COPY . .
RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -a -o seq .

FROM alpine:latest
WORKDIR /gowork/seq
COPY --from=builder /gowork/github.com/spcent/seq/seq ./
COPY --from=builder /gowork/github.com/spcent/seq/config.yml ./
RUN apk add --no-cache tzdata
ENV TZ "Asia/Shanghai"
EXPOSE 8000
ENTRYPOINT ["./seq"]