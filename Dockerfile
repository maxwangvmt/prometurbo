FROM alpine:3.3

RUN apk --update upgrade && apk add ca-certificates && update-ca-certificates
COPY ./prometurbo.linux /bin/prometurbo 

ENTRYPOINT ["/bin/prometurbo"]



