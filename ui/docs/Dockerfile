FROM nginx:alpine3.17

WORKDIR /srv/www

# Nginx
COPY .docker/etc/nginx/nginx.conf /etc/nginx/conf.d/default.conf

COPY . /srv/www

RUN mkdir /srv/www/log
RUN chown -R nginx:nginx /srv/www

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
