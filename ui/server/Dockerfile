FROM php:8.4-zts-bookworm

WORKDIR /srv/www

ARG UID=1000
ARG GID=1000
ENV UID=$UID
ENV GID=$GID
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        autoconf \
        bash \
        ca-certificates \
        cron \
        curl \
        ffmpeg \
        gcc \
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
        libpq-dev \
        make \
        pkg-config \
        supervisor \
        unzip \
        postgresql-client \
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
        pdo_pgsql \
        sockets \
        xsl \
        zip \
    && pecl install redis xdebug \
    && docker-php-ext-enable redis xdebug \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# php-parallax — Go/Rust-semantic concurrency primitive for PHP 8.4 ZTS.
# Build from the upstream tagged source tarball; phpize + configure + make
# install drops parallax.so into the extension dir and the .ini enables it
# unconditionally.
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

# RoadRunner — Go-based PHP application server. Replaces the FPM + nginx
# combo with a single binary that holds warm PHP workers.
ENV RR_VERSION=2025.1.13
RUN curl -fsSL "https://github.com/roadrunner-server/roadrunner/releases/download/v${RR_VERSION}/roadrunner-${RR_VERSION}-linux-amd64.tar.gz" | tar xz -C /tmp \
    && mv "/tmp/roadrunner-${RR_VERSION}-linux-amd64/rr" /usr/local/bin/rr \
    && chmod +x /usr/local/bin/rr \
    && rm -rf "/tmp/roadrunner-${RR_VERSION}-linux-amd64"

COPY --from=composer:2.8 /usr/bin/composer /usr/local/bin/composer

COPY --from=golang:1.21-bookworm /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"
ENV CGO_ENABLED=1

COPY .docker/etc/cron.d/base /etc/cron.d/base
COPY .docker/etc/supervisord/supervisord.conf /etc/supervisor/supervisord.conf
COPY .docker/etc/php/php-dev.ini /usr/local/etc/php/conf.d/50-dev.ini
COPY .docker/etc/php/php-cli.ini /usr/local/etc/php/conf.d/60-dev.ini

RUN mkdir -p /srv/www/var/cache /srv/www/var/log \
    && chmod 644 /etc/cron.d/base

COPY . /srv/www

ENV COMPOSER_ALLOW_SUPERUSER=1

RUN cd /srv/www/microservices/signal-server \
    && rm -f .env \
    && go mod download \
    && go build -o signal-server main.go

RUN cd /srv/www \
    && composer install --no-interaction --optimize-autoloader \
    && chown -R www-data:www-data /srv/www/public /srv/www/var

EXPOSE 80

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]
