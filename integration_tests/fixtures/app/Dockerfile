FROM node:slim

RUN apt-get update && \
    apt-get install -y \
    curl && \
    rm -rf /var/lib/apt/lists/*

RUN npm install -g json http-server

COPY build/containerpilot /bin/containerpilot
COPY containerpilot.json /etc/containerpilot.json
COPY containerpilot-with-coprocess.json /etc/containerpilot-with-coprocess.json
COPY reload-app.sh /reload-app.sh
COPY sensor.sh /sensor.sh
COPY task.sh /task.sh

ENV CONTAINERPILOT=file:///etc/containerpilot.json

# default port, allows us to override in docker-compose and also test
# env var interpolation in the command args
ENV APP_PORT=8000

ENTRYPOINT [ "/bin/containerpilot" ]
