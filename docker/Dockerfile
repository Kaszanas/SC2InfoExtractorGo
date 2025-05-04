# All of the maps and mods downloaded when processing SC2ReSet replays:
FROM kaszanas/sc2reset_maps_mods:latest AS sc2reset_maps_mods

FROM golang:1.24.2-alpine AS build_sc2_info_extractor

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
RUN mkdir -p replays/input
RUN mkdir -p replays/output

COPY --from=build_sc2_info_extractor /app/SC2InfoExtractorGo /app/
# Copying all maps from from SC2ReSet to the maps directory for processing:
# FIXME: Map files are not copied to the container:
COPY --from=sc2reset_maps_mods /sc2reset_maps_mods/maps/cn_maps /app/dependencies/maps/
COPY --from=sc2reset_maps_mods /sc2reset_maps_mods/maps/sc2reset_maps /app/dependencies/maps/
COPY --from=sc2reset_maps_mods /sc2reset_maps_mods/other_dependencies /app/dependencies/other_dependencies/

ENTRYPOINT ["/app/SC2InfoExtractorGo"]
