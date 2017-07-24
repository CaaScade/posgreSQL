#!/bin/sh 

set -x

mkdir -p ~/.aws
touch ~/.aws/config

cat > ~/.aws/config << EOF
[default]
aws_access_key_id=${PUBLIC_KEY}
aws_secret_access_key=${SECRET_KEY}
region=us-east-2
EOF

aws s3 cp s3://sid.caascade/postgresql.tar /postgresql.tar --profile=default

