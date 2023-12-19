FROM golang:latest as build_sc2_info_extractor

WORKDIR /sc2_info_extractor

# Copy Golang dependency definitions:
COPY go.mod go.sum /sc2_info_extractor/

# Get the project dependencies:
RUN --mount=type=cache,target=/go go mod download

COPY . .

# Build the project:
RUN --mount=type=cache,target=/go go build

FROM alpine:latest as final

# libc6-compat is needed for the binary to run on Alpine Linux:
RUN apk add --no-cache ca-certificates libc6-compat

WORKDIR /app

RUN mkdir logs

COPY --from=0 /sc2_info_extractor/SC2InfoExtractorGo /app/

ENTRYPOINT ["/app/SC2InfoExtractorGo"]
