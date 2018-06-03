FROM quay.io/goswagger/swagger:0.14.0
ADD ./templates /templates
ADD ./entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["--help"]
