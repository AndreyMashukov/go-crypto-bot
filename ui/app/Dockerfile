# Build-time stage
FROM node:22-alpine AS build

ENV HOST=0.0.0.0
ENV PORT=3000
ENV TZ=UTC

ARG API_BASE_URL
ENV API_BASE_URL=${API_BASE_URL}
ARG CLIENT_ID
ARG CLIENT_SECRET
ENV CLIENT_ID=${CLIENT_ID}
ENV CLIENT_SECRET=${CLIENT_SECRET}

USER node
RUN mkdir -p /home/node/app
WORKDIR /home/node/app

COPY --chown=node:node package.json yarn.lock ./
RUN yarn install --frozen-lockfile

COPY --chown=node:node . ./
RUN yarn build

# Run-time stage
FROM node:22-alpine

ENV NODE_ENV=production
USER node
RUN mkdir -p /home/node/app
WORKDIR /home/node/app

COPY --chown=node:node nuxt.config.ts package.json yarn.lock ./
COPY --chown=node:node --from=build /home/node/app/.output ./.output

EXPOSE 3000

CMD ["node", ".output/server/index.mjs"]
