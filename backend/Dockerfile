FROM golang AS builder

WORKDIR /app
COPY . .
RUN go build -o main .

FROM alpine
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/main /app
CMD ["/app"]