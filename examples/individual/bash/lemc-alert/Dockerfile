FROM alpine:latest

# Update and install necessary packages using apk
RUN apk update && \
    apk add --no-cache curl shadow

RUN mkdir /lemc

WORKDIR /lemc

COPY lemc/run.sh /lemc/run.sh

RUN chmod +x /lemc/run.sh

CMD ["/bin/sh", "-c", "/lemc/run.sh"]
