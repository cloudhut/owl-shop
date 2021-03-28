############################################################
# Build
############################################################
FROM golang:1.16-alpine as builder
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /app

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o ./bin/owlshop ./cmd

############################################################
# Final Image
############################################################
FROM alpine:3

WORKDIR /app

COPY --from=builder /app/bin/owlshop /app/owlshop

ENTRYPOINT ["./owlshop"]