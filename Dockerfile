FROM golang:alpine

RUN go get github.com/restanrm/twitter
RUN go install github.com/restanrm/listFollowers

ENV TWITTER_KEY
ENV TWITTER_SECRET
ENV TWITTER_USERNAME
ENV NOTIFY_MY_ANDROID_KEY

CMD listFollowers

