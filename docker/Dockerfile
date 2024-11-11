FROM golang:1.23.3-alpine as build_sc2_info_extractor

WORKDIR /app

# Copy Golang dependency definitions:
COPY go.mod go.sum /app/

# Get the project dependencies:
RUN --mount=type=cache,target=/go go mod download

COPY . .

# Build the project:
RUN --mount=type=cache,target=/go go build

FROM alpine:latest as final

RUN apk add --no-cache ca-certificates

WORKDIR /app

RUN mkdir logs

COPY --from=build_sc2_info_extractor /app/SC2InfoExtractorGo /app/

ENTRYPOINT ["/app/SC2InfoExtractorGo"]
