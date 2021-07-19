FROM golang:1.16-alpine AS build
WORKDIR /app
COPY ./ .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o ./simple-web-server
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

FROM scratch
COPY --from=build /app/simple-web-server .
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 8081

ENTRYPOINT ["./simple-web-server"]