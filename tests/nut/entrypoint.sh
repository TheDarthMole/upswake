#! /bin/sh -e

/usr/sbin/upsdrvctl -u root start
/usr/sbin/upsd -u "$USER"
exec /usr/sbin/upsmon -D
