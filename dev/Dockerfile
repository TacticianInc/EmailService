FROM google/golang

WORKDIR /gopath/src
ADD . /gopath/src/

# set go path for 1.6 +
ENV GOPATH /gopath

# get all dependencies
RUN go get github.com/googollee/go-rest

# build application
RUN go build tac/emailservice
RUN go install tac/emailservice

# set env variables

# set port to expose
EXPOSE 8081
CMD []
ENTRYPOINT ["/gopath/bin/emailservice"]