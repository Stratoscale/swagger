FROM quay.io/goswagger/swagger:v0.29.0
ADD ./templates /templates
ADD ./entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["--help"]
