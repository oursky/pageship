FROM alpine:3.17  
RUN apk --no-cache add ca-certificates
COPY controller /controller
ENV PAGESHIP_MIGRATE=true
ENV PAGESHIP_DATABASE_URL=sqlite:///var/pageship/data.db
ENV PAGESHIP_STORAGE_URL=file:///var/pageship/storage?create_dir=true
ENTRYPOINT ["/controller"]
CMD ["start"]
