# Runs the web app's development server (Vite) in a container.
# This image is for development only. Later, for a real server, the app is
# built into static files instead. Learn more: docs/learn/vite.md

FROM node:26.13.0-alpine3.24

WORKDIR /app

# The image's built-in "node" user owns the folder, so the app doesn't run
# as root.
RUN chown node:node /app
USER node

# Install the libraries first, in their own layer, so they are only
# installed again when package.json or package-lock.json change.
# "npm ci" installs exactly the versions in package-lock.json.
COPY --chown=node:node package.json package-lock.json ./
RUN npm ci

COPY --chown=node:node . .

# Vite's development server listens on port 5173.
EXPOSE 5173

CMD ["npm", "run", "dev"]
