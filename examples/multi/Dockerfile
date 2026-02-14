FROM golang:1.23-alpine AS build
WORKDIR /src
COPY main.go .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /api-server main.go

FROM scratch
COPY --from=build /api-server /api-server
EXPOSE 8080
ENTRYPOINT ["/api-server"]
