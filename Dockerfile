# Version 1.0

FROM wolfdeng/alpine

MAINTAINER dongchaofeng@b.360.cn

RUN mkdir /app

ADD bin/log-server /app

ENV LANG en_US.UTF-8
ENV SIMULATOR_ADDR :56555

WORKDIR /app

CMD /app/log-server