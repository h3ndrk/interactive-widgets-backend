FROM node:14.4.0 AS builder

WORKDIR /usr/src
ENV PATH /usr/src/node_modules/.bin:$PATH

COPY frontend/package.json ./
COPY frontend/package-lock.json ./
RUN npm install

COPY frontend .
RUN npm run build

FROM nginx:1.19.0

COPY docker/nginx-frontend/nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /usr/src/build /usr/share/nginx/html
