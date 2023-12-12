FROM golang:latest as build_sc2_info_extractor

WORKDIR /sc2-info-extractor

# Copy Golang dependency definitions:
COPY go.mod go.sum /sc2_info_extractor/

# Get the project dependencies:
RUN --mount=type=cache,target=/go go mod download

COPY . .

# Build the project:
RUN --mount=type=cache,target=/go go build

FROM alpine:latest as final

RUN apk add --no-cache ca-certificates

COPY --from=0 /sc2_info_extractor/SC2InfoExtractorGo .

CMD ["./SC2InfoExtractorGo"]
