FROM alpine:latest

# Update and install necessary packages using apk
RUN apk update && \
    apk add --no-cache curl shadow

RUN mkdir /lemc

WORKDIR /lemc

COPY ./lemc/run.sh .

RUN chmod +x run.sh

CMD ["/lemc/run.sh"]
