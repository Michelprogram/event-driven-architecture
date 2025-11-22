# Use a small, modern Node image
FROM node:20-alpine

# Set workdir
WORKDIR /usr/src/app

# Install deps first (better layer caching)
COPY package*.json ./
ENV NODE_ENV=production
RUN npm ci --omit=dev

# Copy the rest of the app
COPY . .

# Expose the HTTP port used by Frontend-server.js
EXPOSE 8080

# Important: the file is named "Frontend-server.js" (capital F)
# so we run it directly to avoid case-sensitivity issues with "npm start"
CMD ["node", "Frontend-server.js"]