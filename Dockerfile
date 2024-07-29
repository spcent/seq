FROM golang:1.20 as builder
WORKDIR /gowork
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -a -o seq .

FROM alpine:latest
WORKDIR /gowork/seq
COPY --from=builder /gowork/seq ./
COPY --from=builder /gowork/config.yml ./
RUN apk add --no-cache tzdata
ENV TZ "Asia/Shanghai"
EXPOSE 8000
ENTRYPOINT ["./seq"]
