FROM golang:1.16-alpine AS build
WORKDIR /app
ADD main.go /app/main.go
ADD go.mod /app/go.mod
ADD go.sum /app/go.sum
ADD core /app/core/
ADD vendor /app/vendor/
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod vendor -ldflags="-w -s" -o /bin/moxy

FROM alpine:latest
RUN mkdir -p /app
WORKDIR /app
COPY --from=build /bin/moxy .
EXPOSE 8080
CMD [ "./moxy" ]