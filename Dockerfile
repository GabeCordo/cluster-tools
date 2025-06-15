##########################################################################################################
#       Builder Container
##########################################################################################################

FROM golang:1.23.2 AS build-env

WORKDIR /go/src

COPY . .

WORKDIR /go/src/cmd/pipeline-gateway

RUN go mod tidy

# disable cgo so the binary can be brought to a smaller container
ENV CGO_ENABLED=0
RUN go build -o pipeline-gateway
RUN ./pipeline-gateway init
RUN ./pipeline-gateway doctor

##########################################################################################################
#       Production Container
##########################################################################################################

FROM gcr.io/distroless/static-debian12

COPY --from=build-env /go/src/cmd/pipeline-gateway/pipeline-gateway /
COPY --from=build-env /root/.cache/flock /root/.cache/flock

EXPOSE 8136
EXPOSE 8137

ENTRYPOINT ["/pipeline-gateway", "start"]