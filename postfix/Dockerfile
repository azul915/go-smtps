FROM ubuntu:trusty

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && \
    apt-get -y install --no-install-recommends \
    supervisor postfix sasl2-bin \
    opendkim opendkim-tools dnsutils curl

COPY       supervisord.conf /etc/supervisord.conf
COPY       rsyslog.conf /etc/rsyslog.conf
COPY       run.sh /run.sh
RUN        chmod +x /run.sh

USER       root
WORKDIR    /tmp

EXPOSE     25
EXPOSE     587
ENTRYPOINT ["/run.sh"]
