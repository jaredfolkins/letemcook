FROM ubuntu:20.04
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y curl \
    sudo

RUN mkdir /lemc

WORKDIR /lemc

COPY lemc/script.js /lemc/script.js
COPY lemc/run.sh /lemc/run.sh

RUN chmod +x /lemc/run.sh

CMD ["/bin/bash", "-c", "/lemc/run.sh"]
