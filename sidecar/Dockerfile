FROM debian:latest

RUN mkdir /resources

RUN apt-get update && apt-get install -y postgresql-client-9.6

ADD posgres-sidecar .

ADD resourcev1/ /resources/

ENTRYPOINT ["/posgres-sidecar"]
