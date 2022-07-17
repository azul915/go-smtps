#!/bin/bash
DOMAIN_DEFAULT=postfix.example.com
DOMAIN=${DOMAIN:-$DOMAIN_DEFAULT}

# while [[ "$(curl -s -o /dev/null -w ''%{http_code}'' mailcatcher:1080)" != "200" ]]; do sleep 3; echo waiting for mailcatcher...; done
while [[ "$(curl -s -o /dev/null -w ''%{http_code}'' maildev:1080)" != "200" ]]; do sleep 3; echo waiting for maildev...; done

echo "******************************"
echo "**** POSTFIX STARTING UP *****"
echo "******************************"

echo "DOMAIN: ${DOMAIN}"
# export RELAYHOSTIP=$(dig mailcatcher +short)
export RELAYHOSTIP=$(dig maildev +short)
echo "RELAY : ${RELAYHOSTIP}:${RELAYHOSTPORT}"

# Make and reown postfix folders
mkdir -p /var/spool/postfix/ && mkdir -p /var/spool/postfix/pid
chown root: /var/spool/postfix/
chown root: /var/spool/postfix/pid

# Disable SMTPUTF8, because libraries (ICU) are missing in alpine
postconf -e smtputf8_enable=no

# Update aliases database. It's not used, but postfix complains if the .db file is missing
postalias /etc/postfix/aliases

# Disable local mail delivery
postconf -e mydestination=
# Don't relay for any domains
postconf -e relay_domains=
# Reject invalid HELOs
postconf -e smtpd_delay_reject=yes
postconf -e smtpd_helo_required=yes
postconf -e "smtpd_helo_restrictions=permit_mynetworks,reject_invalid_helo_hostname,permit"

postconf -e "smtpd_tls_loglevel=3"
# TLS settings
postconf -e "smtp_tls_security_level=may"
postconf -e "smtp_tls_note_starttls_offer=no"
postconf -e "smtp_tls_CApath=/etc/ssl/certs"

postconf -e myhostname="$DOMAIN"
postconf -e mydomain="$DOMAIN"
postconf -e relayhost=[$RELAYHOSTIP]:$RELAYHOSTPORT
postconf -e "smtp_sasl_auth_enable=no"

postconf -e "smtp_host_lookup=native"
#postconf -e "disable_dns_lookups=yes"

postconf -e smtpd_sasl_security_options=noanonymous
############
# SASL SUPPORT FOR CLIENTS
# The following options set parameters needed by Postfix to enable
# Cyrus-SASL support for authentication of mail clients.
############
# /etc/postfix/main.cf
postconf -e smtpd_sasl_auth_enable=yes
postconf -e smtpd_sasl_local_domain=$DOMAIN
postconf -e broken_sasl_auth_clients=yes
postconf -e smtpd_recipient_restrictions=permit_sasl_authenticated,reject_unauth_destination
# smtpd.conf
cat >> /etc/postfix/sasl/smtpd.conf <<EOF
pwcheck_method: auxprop
auxprop_plugin: sasldb
mech_list: PLAIN LOGIN CRAM-MD5 DIGEST-MD5 NTLM
EOF
# sasldb2
echo SMTP_USER: $SMTP_USER, SMTP_PASSWORD: $SMTP_PASSWORD
echo $SMTP_PASSWORD | saslpasswd2 -p -c -u $DOMAIN $SMTP_USER
chown postfix.sasl /etc/sasldb2
#chown postfix.postfix /etc/sasldb2
#postfixが参照するためハードリンクする http://kt-hiro.hatenablog.com/entry/20120318/1332023507
ln /etc/sasldb2 /var/spool/postfix/etc/sasldb2
# vi /etc/postfix/master.cf
# -o smtpd_sasl_auth_enable=yes

# Since we are behind closed doors, let's just permit all relays.
postconf -e "smtpd_relay_restrictions=permit"

# Use 587 (submission)
sed -i -r -e 's/^#submission/submission/' /etc/postfix/master.cf

echo "- Staring rsyslog and postfix"
exec supervisord -c /etc/supervisord.conf
