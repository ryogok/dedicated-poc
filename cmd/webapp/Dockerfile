### Build app
FROM golang:1.19 AS build

WORKDIR /app

COPY . .
RUN go mod download
RUN go build -o /webapp ./cmd/webapp

### Copy app to runtime container
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /webapp /webapp

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/webapp"]