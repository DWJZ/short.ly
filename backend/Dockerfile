FROM golang:1.26.1-alpine AS dev

WORKDIR /app
RUN apk add --no-cache git ca-certificates && update-ca-certificates

COPY go.mod ./
RUN go mod download

RUN go install github.com/air-verse/air@latest

COPY . .
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]


FROM golang:1.26.1-alpine AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api


FROM gcr.io/distroless/static-debian12:nonroot AS prod

WORKDIR /
COPY --from=build /out/api /api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]
