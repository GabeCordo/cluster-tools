FROM golang:alpine3.18

WORKDIR /home/app

COPY . .

WORKDIR /home/app/cmd/cluster-tools

RUN go install
RUN ctgate init
RUN ctgate doctor

EXPOSE 8136
EXPOSE 8137

ENTRYPOINT ["cluster-tools", "start"]