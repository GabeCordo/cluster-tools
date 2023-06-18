#######################################
# use a ubuntu base image as an entrypoint
#######################################

FROM ubuntu:latest
WORKDIR /home/

RUN apt-get update \
    && apt-get install -y wget \
    && apt-get install -y git \
    && apt-get install -y curl

#######################################
# Networking Configuraiton
#######################################

# HTTP API Port for ETL Service
EXPOSE 8136

#######################################
# setup lanugage dependencies
#######################################

ARG GO_VERSION="go1.20.2.linux-amd64.tar.gz"

RUN /usr/bin/wget "https://golang.org/dl/$GO_VERSION"
RUN curl -sL https://golang.org/dl/ | grep -A 5 -w $GO_VERSION
RUN tar -C /usr/local -xzf $GO_VERSION
RUN chown -R root:root /usr/local/go/bin
RUN rm $GO_VERSION
ENV PATH="${PATH}:/usr/local/go/bin"

#######################################
# Settup ETL Repository
#######################################

WORKDIR /home/

# setup for ETL Sentiment
RUN mkdir /etc/etl/
RUN mkdir /etc/etl/configs
COPY .bin/configs/config.etl.yaml /etc/etl/configs/
RUN mkdir /etc/etl/modules

COPY . /home/etlsrc/

WORKDIR /home/etlsrc/

RUN go clean -modcache
RUN go mod tidy
RUN go build -o /home/etl
RUN chmod +x /home/etl
ENV PATH="${PATH}:/home/"

RUN rm -rf /home/etlsrc
RUN rm -d /home/etlsrc

#######################################
# ETL Runtime Arguments
#######################################

ENV CONFIG_PATH=/etc/etl/configs/config.etl.yaml
ENV MODULES_PATH=/etc/etl/modules/

#######################################
# Initialization and Startup
#######################################

WORKDIR /home/

ENTRYPOINT etl --config $CONFIG_PATH --modules $MODULES_PATH
