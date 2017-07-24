FROM debian:latest

RUN mkdir /resources

RUN apt-get update && apt-get install -y postgresql-client-9.6 \
						openssl \
						curl \
						awscli

ADD posgres-sidecar .

ADD resourcev1/ /resources/
ADD get-s3-backup.sh /
ADD put-s3-backup.sh /

ENTRYPOINT ["/posgres-sidecar"]
