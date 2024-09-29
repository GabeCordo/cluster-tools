FROM golang:alpine3.18

WORKDIR /home/app

COPY . .

WORKDIR /home/app/cmd/ctgate

RUN go install
RUN ctgate init
RUN ctgate doctor

EXPOSE 8136
EXPOSE 8137

ENTRYPOINT ["ctgate", "start"]