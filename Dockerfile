# Version 1.0

FROM wolfdeng/alpine

MAINTAINER dongchaofeng@b.360.cn

RUN mkdir /app

ADD bin/monitor /app

ENV LANG en_US.UTF-8
ENV SIMULATOR_ADDR :56555
ENV SIMULATOR_LOG ./log/monitor.log

WORKDIR /app

CMD /app/monitor