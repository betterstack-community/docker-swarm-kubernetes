
FROM golang:1.19

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
COPY config/db.go ./config/db.go
COPY handler/handlers.go ./handler/handlers.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/main ./
EXPOSE 8081

# Run
ENTRYPOINT ["/out/main"]