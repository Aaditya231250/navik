FROM node:18-alpine

# Install dependencies
RUN apk add --no-cache git

# Install Expo CLI globally
RUN npm install -g expo-cli

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy the rest of the application
COPY . .

# Expose Expo and Metro bundler ports
EXPOSE 19000 19001 19002 8081

# Start Expo development server
CMD ["npx", "expo", "start", "--host", "lan"]
