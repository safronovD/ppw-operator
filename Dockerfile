#Building image
FROM golang:1.13 AS builder

WORKDIR /workspace

COPY go.mod go.mod
RUN go mod download

COPY cmd/main.go main.go
COPY pkg/api api/
COPY pkg/controllers controllers/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

#Running image
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/manager .
USER nonroot:nonroot
ENTRYPOINT ["/manager"]
