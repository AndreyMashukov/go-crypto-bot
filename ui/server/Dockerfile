FROM php:8.4-zts-bookworm

WORKDIR /srv/www

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        autoconf \
        ca-certificates \
        curl \
        gcc \
        git \
        libffi-dev \
        libicu-dev \
        libonig-dev \
        libxml2-dev \
        libxslt1-dev \
        libzip-dev \
        make \
        pkg-config \
        unzip \
        zlib1g-dev \
    && docker-php-ext-install -j"$(nproc)" \
        bcmath \
        iconv \
        intl \
        mbstring \
        opcache \
        pcntl \
        sockets \
        xsl \
        zip \
    && pecl install redis \
    && docker-php-ext-enable redis \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# php-parallax — Go/Rust-semantic concurrency primitive for PHP 8.4 ZTS.
ENV PHP_PARALLAX_VERSION=0.2.1
RUN curl -fsSL "https://github.com/AndreyMashukov/php-parallax/archive/refs/tags/v${PHP_PARALLAX_VERSION}.tar.gz" | tar xz -C /tmp \
    && cd "/tmp/php-parallax-${PHP_PARALLAX_VERSION}" \
    && phpize \
    && ./configure --enable-parallax \
    && make -j"$(nproc)" \
    && make install \
    && echo "extension=parallax.so" > /usr/local/etc/php/conf.d/parallax.ini \
    && php -m | grep parallax \
    && cd / && rm -rf "/tmp/php-parallax-${PHP_PARALLAX_VERSION}"

# RoadRunner — Go-based PHP application server.
ENV RR_VERSION=2025.1.13
RUN curl -fsSL "https://github.com/roadrunner-server/roadrunner/releases/download/v${RR_VERSION}/roadrunner-${RR_VERSION}-linux-amd64.tar.gz" | tar xz -C /tmp \
    && mv "/tmp/roadrunner-${RR_VERSION}-linux-amd64/rr" /usr/local/bin/rr \
    && chmod +x /usr/local/bin/rr \
    && rm -rf "/tmp/roadrunner-${RR_VERSION}-linux-amd64"

COPY --from=composer:2.8 /usr/bin/composer /usr/local/bin/composer

COPY .docker/etc/php/php-dev.ini /usr/local/etc/php/conf.d/50-dev.ini
COPY .docker/etc/php/php-cli.ini /usr/local/etc/php/conf.d/60-dev.ini

COPY . /srv/www

ENV COMPOSER_ALLOW_SUPERUSER=1

RUN mkdir -p /srv/www/var/cache /srv/www/var/log \
    && composer install --no-interaction --optimize-autoloader --no-scripts \
    && php bin/console cache:clear --no-warmup \
    && chown -R www-data:www-data /srv/www/public /srv/www/var

EXPOSE 80

CMD ["/usr/local/bin/rr", "serve", "-c", "/srv/www/.rr.yaml"]
