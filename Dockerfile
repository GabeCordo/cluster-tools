##########################################################################################################
#       Builder Container
##########################################################################################################

FROM golang:1.24.6 AS build-env

# If you want to force cache invalidation after changing a secret value,
# you can pass a build argument with an arbitrary value that you also change
# when changing the secret. Build arguments do result in cache invalidation.
# ref. https://docs.docker.com/build/cache/invalidation/
ARG CACHEBUST

# Add the .netrc file to the builder container
RUN touch $HOME/.netrc && \
chmod 600 $HOME/.netrc

# Store the GitHub Username and Access Token (granted in the GitHub Organization)
# in a .netrc file so Golan can use HTTPS authentication with Git.
#
# NOTE: If the secrets change, we need to push CACHEBUST=1 otherwise CACHEBUST=0
#       will always be true.
RUN --mount=type=secret,id=username  \
    --mount=type=secret,id=token  \
    echo "machine github.com login $(cat /run/secrets/username) password $(cat /run/secrets/token)" > $HOME/.netrc

WORKDIR /go/src
COPY . .

WORKDIR /go/src/cmd/flock

# Private Imports Will Fall Under ForitifiedCode
ENV GOPRIVATE=github.com/FortifiedCode

RUN go mod tidy

# disable cgo so the binary can be brought to a smaller container
ENV CGO_ENABLED=0
RUN go build -o flock --tags=production
RUN ./flock init
RUN ./flock doctor

##########################################################################################################
#       Production Container
##########################################################################################################

FROM gcr.io/distroless/static-debian12

COPY --from=build-env /go/src/cmd/flock/flock /
COPY --from=build-env /root/.cache/flock /root/.cache/flock

EXPOSE 8136
EXPOSE 8137

ENTRYPOINT ["/flock", "start"]