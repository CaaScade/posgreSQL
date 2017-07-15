FROM debian:latest

ADD posgres-sidecar .

RUN mkdir /resources

RUN apt-get update && apt-get install -y postgresql-client

ADD resources /resources/

ENTRYPOINT ["/posgres-sidecar"]
