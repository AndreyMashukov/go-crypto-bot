FROM php:8.3-fpm-bookworm

WORKDIR /srv/www

ARG UID=1000
ARG GID=1000
ENV UID=$UID
ENV GID=$GID
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        bash \
        ca-certificates \
        cron \
        curl \
        ffmpeg \
        git \
        libcap2-bin \
        libffi-dev \
        libfreetype6-dev \
        libicu-dev \
        libjpeg62-turbo-dev \
        libonig-dev \
        libpng-dev \
        libwebp-dev \
        libxml2-dev \
        libxslt1-dev \
        libzip-dev \
        nginx \
        pkg-config \
        supervisor \
        unzip \
        default-mysql-client \
        zlib1g-dev \
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
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY --from=composer:2 /usr/bin/composer /usr/local/bin/composer

COPY --from=golang:1.21-bookworm /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"
ENV CGO_ENABLED=1

COPY .docker/etc/cron.d/base /etc/cron.d/base
COPY .docker/etc/supervisord/supervisord.conf /etc/supervisor/supervisord.conf
COPY .docker/etc/php/php-dev.ini /usr/local/etc/php/conf.d/50-dev.ini
COPY .docker/etc/php/php-cli.ini /usr/local/etc/php/conf.d/60-dev.ini
COPY .docker/etc/php/php-fpm-pool.conf /usr/local/etc/php-fpm.conf
COPY .docker/etc/nginx/nginx.conf /etc/nginx/conf.d/default.conf
RUN mkdir -p /run/nginx /run/php /srv/www/var/cache /srv/www/var/log \
    && chmod 644 /etc/cron.d/base

COPY . /srv/www

ENV COMPOSER_ALLOW_SUPERUSER=1

RUN cd /srv/www/microservices/signal-server \
    && rm -f .env \
    && go mod download \
    && go build -o signal-server main.go

RUN cd /srv/www \
    && composer install --no-interaction --optimize-autoloader --no-audit \
    && chown -R www-data:www-data /srv/www/public /srv/www/var

EXPOSE 80

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]
