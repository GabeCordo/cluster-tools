##########################################################################################################
#       Builder Container
##########################################################################################################

FROM golang:1.23.2 AS build-env

WORKDIR /go/src

COPY . .

WORKDIR /go/src/cmd/ctgate

RUN go mod tidy

ENV CGO_ENABLED=0
RUN go build -o ctgate
RUN ./ctgate init
RUN ./ctgate doctor

##########################################################################################################
#       Production Container
##########################################################################################################

FROM gcr.io/distroless/static-debian12

COPY --from=build-env /go/src/cmd/ctgate/ctgate /
COPY --from=build-env /root/.cache/cluster.tools /root/.cache/cluster.tools

EXPOSE 8136
EXPOSE 8137

ENTRYPOINT ["/ctgate", "start"]