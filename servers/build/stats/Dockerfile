# 01 - build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY . .

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN go build -o stats ./cmd/stats

# 02 - run
FROM scratch

COPY --from=builder /app /

EXPOSE 8084

CMD [ "/stats" ]
