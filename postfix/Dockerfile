FROM ubuntu:trusty

# Set noninteractive mode for apt-get
ENV DEBIAN_FRONTEND noninteractive

# Update
RUN apt-get update

# Start editing
# Install package here for cache
RUN apt-get -y install supervisor postfix sasl2-bin opendkim opendkim-tools

COPY       supervisord.conf /etc/supervisord.conf
COPY       rsyslog.conf /etc/rsyslog.conf
COPY       run.sh /run.sh
RUN        chmod +x /run.sh

# Run supervisord
USER       root
WORKDIR    /tmp

EXPOSE     25
EXPOSE     587
ENTRYPOINT ["/run.sh"]
