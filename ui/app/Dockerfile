# Build-time stage — Nuxt 4 / Vue 3.5 / Vuetify 3 / Pinia
FROM node:22-alpine AS build

ENV HOST=0.0.0.0
ENV PORT=3000
ENV TZ=UTC

WORKDIR /app

COPY package.json ./
RUN corepack enable && corepack prepare yarn@1.22.22 --activate \
    && yarn install --network-timeout 600000

COPY . ./
RUN yarn build

# Run-time stage
FROM node:22-alpine

ENV NODE_ENV=production
WORKDIR /app

COPY --from=build /app/.output ./.output

EXPOSE 3000

CMD ["node", ".output/server/index.mjs"]
