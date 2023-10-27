FROM golang:alpine3.18

WORKDIR /home/app

COPY . .
RUN go install
RUN mango init
RUN mango doctor

EXPOSE 8136
EXPOSE 8137

ENTRYPOINT ["mango", "start"]