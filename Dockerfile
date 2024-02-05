FROM alpine
RUN apk add --no-cache bash ca-certificates
COPY s3-index-generator /usr/bin/s3-index-generator
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /usr/bin/s3-index-generator /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
