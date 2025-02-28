#Build stage
FROM golang:1.22.5-alpine3.19 AS builder
WORKDIR /app
COPY . .
RUN go build -o main ./cmd/spyCatAgency/main.go

#Run stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/.env .
COPY --from=builder /app/main .
RUN mkdir migrations
COPY --from=builder /app/internal/infrastructure/database/migrations /app/migrations






EXPOSE 8080
CMD [ "/app/main" ]