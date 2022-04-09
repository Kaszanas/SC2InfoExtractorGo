FROM golang:1.15-alpine

RUN mkdir /sc2-info-extractor
WORKDIR /sc2-info-extractor
COPY go.mod .
COPY go.sum .
RUN --mount=type=cache,target=/go go mod download

COPY . .
RUN rm -f SC2InfoExtractorGo
RUN --mount=type=cache,target=/go go build

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=0 /sc2-info-extractor/SC2InfoExtractorGo .
CMD ["./SC2InfoExtractorGo"]
