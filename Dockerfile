FROM quay.io/goswagger/swagger:0.13.0
ADD ./templates /templates
ADD ./entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["--help"]
