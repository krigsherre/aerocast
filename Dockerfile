FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=unknown

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commitSHA=${COMMIT}" \
    -o /aerocastd ./cmd/aerocastd/

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /aerocastd /aerocastd
COPY configs/default.yaml /etc/aerocast/config.yaml

EXPOSE 9100 9101 9102

USER nonroot:nonroot

ENTRYPOINT ["/aerocastd"]
CMD ["-config", "/etc/aerocast/config.yaml"]
