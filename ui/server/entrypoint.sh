#!/bin/sh

chown -R www-data:www-data /srv/www/public
chown -R www-data:www-data /srv/www/var/log

exec "$@"
