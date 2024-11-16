# All of the maps and mods downloaded when processing SC2ReSet replays:
FROM kaszanas/sc2reset_maps_mods:latest AS sc2reset_maps_mods

FROM golang:1.23.3-alpine AS build_sc2_info_extractor

WORKDIR /app

# Copy Golang dependency definitions:
COPY go.mod go.sum /app/

# Get the project dependencies:
RUN --mount=type=cache,target=/go go mod download

COPY . .

# Build the project:
RUN --mount=type=cache,target=/go go build

FROM alpine:latest AS final

RUN apk add --no-cache ca-certificates

WORKDIR /app

RUN mkdir logs

COPY --from=build_sc2_info_extractor /app/SC2InfoExtractorGo /app/
# Copying all maps from from SC2ReSet to the maps directory for processing:
COPY --from=sc2reset_maps_mods /sc2reset_maps_mods/maps/cn_maps/ /app/maps/
COPY --from=sc2reset_maps_mods /sc2reset_maps_mods/maps/sc2reset_maps/ /app/maps/

ENTRYPOINT ["/app/SC2InfoExtractorGo"]
