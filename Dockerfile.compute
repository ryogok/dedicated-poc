### Build app
FROM golang:1.19 AS build

WORKDIR /app

COPY . .
RUN go mod download
RUN go build -o /compute ./cmd/compute

### Copy app to runtime container
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /compute /compute

EXPOSE 8081

USER nonroot:nonroot

ENTRYPOINT ["/compute"]