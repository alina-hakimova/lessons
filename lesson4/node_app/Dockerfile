FROM node:20-alpine
WORKDIR /app
COPY index.js .
USER node
ENV NODE_ENV=production
EXPOSE 8002
HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD wget -qO- http://localhost:8002/ || exit 1
CMD ["node", "index.js"]
