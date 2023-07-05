FROM golang:latest AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ap-bot

FROM alpine:latest
WORKDIR /src
COPY --from=builder /src/ap-bot /bin/
ENTRYPOINT ["ap-bot"]
