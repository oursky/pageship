FROM alpine:3.17  
RUN apk --no-cache add ca-certificates
COPY pageship /usr/bin/pageship
WORKDIR /var/pageship
ENTRYPOINT ["/usr/bin/pageship"]
CMD ["serve", "/var/pageship"]
