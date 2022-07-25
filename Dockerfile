FROM golang:1.18 AS build
COPY / /src
WORKDIR /src
RUN --mount=type=cache,target=/go/pkg --mount=type=cache,target=/root/.cache/go-build make build

FROM alpine:3.16.0 AS base
RUN apk add --no-cache ca-certificates 
RUN adduser -D acorn
# USER acorn commented out so that i can easily get into the container and do stuff to it
ENTRYPOINT ["/usr/local/bin/acorn-dns"]
CMD ["api-server"]

FROM base
COPY --from=build /src/bin/acorn-dns /usr/local/bin/acorn-dns