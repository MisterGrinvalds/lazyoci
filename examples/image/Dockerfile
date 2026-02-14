ARG GO_VERSION=1.23
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
COPY main.go .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /hello-server main.go

FROM scratch
COPY --from=build /hello-server /hello-server
EXPOSE 8080
ENTRYPOINT ["/hello-server"]
