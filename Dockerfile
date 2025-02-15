FROM golang:1.23.6-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./server.go

FROM alpine:3.17
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=build /app/server /app/server
USER appuser
EXPOSE 8080
CMD ["./server"]