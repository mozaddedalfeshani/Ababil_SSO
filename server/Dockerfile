FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/serve ./cmd/serve
RUN CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate
RUN CGO_ENABLED=0 go build -o /out/retention ./cmd/retention

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/serve /serve
COPY --from=build /out/migrate /migrate
COPY --from=build /out/retention /retention
USER nonroot:nonroot
ENTRYPOINT ["/serve"]
