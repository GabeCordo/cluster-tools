FROM golang:alpine3.18

WORKDIR /home/app

COPY . .
RUN go install
RUN cluster-tools init
RUN cluster-tools doctor

EXPOSE 8136
EXPOSE 8137

ENTRYPOINT ["cluster-tools", "start"]