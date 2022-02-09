FROM alpine

WORKDIR /
COPY ./bin/colonies /bin

CMD ["colonies", "server", "start"]
