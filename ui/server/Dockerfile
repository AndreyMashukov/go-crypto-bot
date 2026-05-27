FROM php:8.3-fpm-alpine

WORKDIR /srv/www

ARG UID=1000
ARG GID=1000
ENV UID=$UID
ENV GID=$GID

# Build dependencies + runtime libs.
RUN apk add --no-cache \
        bash \
        ca-certificates \
        curl \
        dcron \
        ffmpeg \
        freetype \
        freetype-dev \
        gcc \
        git \
        icu-dev \
        jpeg-dev \
        libc-dev \
        libcap \
        libjpeg-turbo-dev \
        libpng \
        libpng-dev \
        libwebp \
        libwebp-dev \
        libwebp-tools \
        libxml2-dev \
        libxslt-dev \
        linux-headers \
        mysql-client \
        nginx \
        oniguruma-dev \
        supervisor \
        unzip \
    && docker-php-ext-configure gd --with-freetype --with-jpeg --with-webp \
    && docker-php-ext-install -j"$(nproc)" \
        bcmath \
        gd \
        iconv \
        intl \
        mbstring \
        opcache \
        pcntl \
        pdo_mysql \
        sockets \
        xsl \
        zip \
    && pecl install redis xdebug \
    && docker-php-ext-enable redis xdebug \
    && apk del --no-cache gcc libc-dev linux-headers

# Composer 2 from official image — pinned, deterministic.
COPY --from=composer:2 /usr/bin/composer /usr/local/bin/composer

# Go toolchain for in-image microservice build.
COPY --from=golang:1.21-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"
ENV CGO_ENABLED=1

# cron + supervisor + nginx + php-fpm wiring.
COPY .docker/etc/cron.d/base /etc/crontabs/root
COPY .docker/etc/supervisord/supervisord.conf /etc/supervisor/supervisord.conf
COPY .docker/etc/php/php-dev.ini /usr/local/etc/php/conf.d/50-dev.ini
COPY .docker/etc/php/php-cli.ini /usr/local/etc/php/conf.d/60-dev.ini
COPY .docker/etc/php/php-fpm-pool.conf /usr/local/etc/php-fpm.conf
COPY .docker/etc/nginx/nginx.conf /etc/nginx/conf.d/default.conf
RUN mkdir -p /run/nginx /run/php /srv/www/var/cache /srv/www/var/log \
    && chmod 644 /etc/crontabs/*

# Code + dependencies.
COPY . /srv/www

ENV COMPOSER_ALLOW_SUPERUSER=1

# Build the signal-server microservice in-image (legacy parity with PHP 7.4 image).
RUN cd /srv/www/microservices/signal-server \
    && rm -f .env \
    && go mod download \
    && go build -o signal-server main.go

RUN cd /srv/www \
    && composer install --no-interaction --optimize-autoloader \
    && chown -R www-data:www-data /srv/www/public /srv/www/var

EXPOSE 80

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]
