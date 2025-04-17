# 🚗 NAVIK - On-Demand Ride Booking Platform 🚗

<div align="center">

```
 ███╗   ██╗ █████╗ ██╗   ██╗██╗██╗  ██╗
 ████╗  ██║██╔══██╗██║   ██║██║██║ ██╔╝
 ██╔██╗ ██║███████║██║   ██║██║█████╔╝ 
 ██║╚██╗██║██╔══██║╚██╗ ██╔╝██║██╔═██╗ 
 ██║ ╚████║██║  ██║ ╚████╔╝ ██║██║  ██╗
 ╚═╝  ╚═══╝╚═╝  ╚═╝  ╚═══╝  ╚═╝╚═╝  ╚═╝


```

<a href="#features"><img src="https://img.shields.io/badge/FEATURES-4A154B?style=for-the-badge" height="30"></a>
<a href="#tech-stack"><img src="https://img.shields.io/badge/TECH_STACK-FF5733?style=for-the-badge" height="30"></a>
<a href="#installation"><img src="https://img.shields.io/badge/INSTALLATION-007ACC?style=for-the-badge" height="30"></a>

</div>
## 👥 Authors

<div align="center">
  
  | Name | Profile |
  |:---:|:---:|
  | Naman Soni | [LinkedIn](https://www.linkedin.com/in/naman-soni-a46931290) |
  | Kishore Vishal | [LinkedIn](https://www.linkedin.com/in/kishore-vishal/) |
  | Aaditya Jain | [LinkedIn](https://www.linkedin.com/in/aaditya-jain-7b85b028b/) |
  | Japneet Singh | [LinkedIn](https://www.linkedin.com/in/japneet-singh-81347b28b/) |
  | William Samuel | [LinkedIn]|
  
</div>
<hr>

## 📊 Overview 📊

NAVIK is a comprehensive ride-hailing platform designed to provide on-demand transportation services through an intuitive mobile application interface. The platform replicates core functionalities of existing ride-hailing services while introducing enhancements for scalability, efficiency, and user experience.


## ✨ Features

```
╔═══════════════════════════════════════════════════════════════════════╗
║  • Real-Time Ride Matching with intelligent driver allocation          ║
║  • Live GPS Tracking with accurate ETA updates                         ║
║  • RideBooking and Tracking of the Driver                              ║
╚═══════════════════════════════════════════════════════════════════════╝
```

## 🚀 Tech Stack

<div align="center">

[![Frontend](https://img.shields.io/badge/Frontend-React_Native-61DAFB?style=for-the-badge&logo=react&logoColor=white)](https://reactnative.dev/)
[![Backend](https://img.shields.io/badge/Backend-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org/)
[![Database](https://img.shields.io/badge/Database-DynamoDB-4053D6?style=for-the-badge&logo=amazon-dynamodb&logoColor=white)](https://aws.amazon.com/dynamodb/)
[![Messaging](https://img.shields.io/badge/Messaging-Apache_Kafka-231F20?style=for-the-badge&logo=apache-kafka&logoColor=white)](https://kafka.apache.org/)
[![Maps](https://img.shields.io/badge/Maps-HereMaps-00AFAA?style=for-the-badge&logo=here&logoColor=white)](https://www.here.com/)
[![Deployment](https://img.shields.io/badge/Deployment-Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)

</div>

## 🌟 Key Components

<div align="center">

```
     ┌─────────────┐      ┌──────────────┐      ┌────────────┐      ┌──────────────┐
     │Frontend Apps│      │Backend       │      │Real-time   │      │Payment       │
     │(Rider/Driver│◄────►│Microservices │◄────►│Tracking    │◄────►│Processing    │
     └─────────────┘      └──────────────┘      └────────────┘      └──────────────┘
          ▲                     ▲                    ▲                    ▲
          │                     │                    │                    │
          ▼                     ▼                    ▼                    ▼
     ┌─────────────┐      ┌──────────────┐      ┌────────────┐      ┌──────────────┐
     │User         │      │Route         │      │Safety      │      │Analytics     │
     │Management   │◄────►│Optimization  │◄────►│Features    │◄────►│& Reporting   │
     └─────────────┘      └──────────────┘      └────────────┘      └──────────────┘
```

</div>

## 💡 System Architecture

NAVIK follows a microservices architecture pattern to ensure scalability and maintainability:

- **Frontend Applications**: Separate React Native apps for riders and drivers
- **Backend Services**: Go-based microservices for user management, ride matching, payment processing
- **Data Storage**: DynamoDB for structured and unstructured data management
- **Message Broker**: Apache Kafka for real-time driver location updates
- **Load Balancing**: Zookeeper for high availability and efficient request distribution
- **Containerization**: Docker for consistent deployment across environments

## 📱 User Interfaces

### Passenger App Features
- Intuitive ride booking with address search
- Real-time driver tracking
- Multiple payment options
- Ride history and receipts
- Safety features including emergency button and ride sharing


## 🔐 Security & Safety

- End-to-end encryption for sensitive data
- Multi-factor authentication for drivers and admins
- PCI DSS compliance for payment processing
- Emergency button with location sharing
- Driver verification and background checks
- Ride sharing with trusted contacts

## 📊 Performance Metrics

- API response time under 300ms
- Support for 1,000+ concurrent ride requests
- Driver matching within 10 seconds
- GPS location updates every 5 seconds
- 95% system availability with failover mechanisms

## 📥 Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/navik.git

# Navigate to project directory
cd navik

# Install dependencies
npm install

# Set up environment variables
cp .env.example .env
nano .env  # Edit with your configuration

# Run the development server
npm run dev
```

## 🔧 Configuration

Create a `.env` file with the following configurations:

```
# API Keys
HERE_MAPS_API_KEY=your_here_maps_api_key
STRIPE_API_KEY=your_stripe_api_key

# Database
DYNAMODB_REGION=your_aws_region
DYNAMODB_ACCESS_KEY=your_aws_access_key
DYNAMODB_SECRET_KEY=your_aws_secret_key

# Kafka Configuration
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_LOCATION=driver-location-updates

# Server Configuration
PORT=8080
NODE_ENV=development
```

## 📝 Usage

1. Register as a rider or driver
2. Complete profile setup and verification
3. Enable location services
4. For riders: Enter pickup/drop-off locations and book a ride
5. For drivers: Accept ride requests and follow navigation

## 📚 API Documentation

API documentation is available at `/api/docs` when running the development server, or visit our [API Documentation](https://api.navik.example.com/docs).

## 💼 Business Rules

- Riders must have a verified payment method
- Drivers can only accept rides in their registered city
- Cancellation fees apply after driver assignment
- Drivers must maintain minimum 4.0 star rating
- Surge pricing applies during high demand periods
- Refunds processed within 5-7 business days



## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

```
████████╗██╗  ██╗ █████╗ ███╗   ██╗██╗  ██╗███████╗    ███████╗ ██████╗ ██████╗     ██╗   ██╗██╗███████╗██╗████████╗██╗███╗   ██╗ ██████╗ ██╗
╚══██╔══╝██║  ██║██╔══██╗████╗  ██║██║ ██╔╝██╔════╝    ██╔════╝██╔═══██╗██╔══██╗    ██║   ██║██║██╔════╝██║╚══██╔══╝██║████╗  ██║██╔════╝ ██║
   ██║   ███████║███████║██╔██╗ ██║█████╔╝ ███████╗    █████╗  ██║   ██║██████╔╝    ██║   ██║██║███████╗██║   ██║   ██║██╔██╗ ██║██║  ███╗██║
   ██║   ██╔══██║██╔══██║██║╚██╗██║██╔═██╗ ╚════██║    ██╔══╝  ██║   ██║██╔══██╗    ╚██╗ ██╔╝██║╚════██║██║   ██║   ██║██║╚██╗██║██║   ██║╚═╝
   ██║   ██║  ██║██║  ██║██║ ╚████║██║  ██╗███████║    ██║     ╚██████╔╝██║  ██║     ╚████╔╝ ██║███████║██║   ██║   ██║██║ ╚████║╚██████╔╝██╗
   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝    ╚═╝      ╚═════╝ ╚═╝  ╚═╝      ╚═══╝  ╚═╝╚══════╝╚═╝   ╚═╝   ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚═╝
```

⭐ Don't forget to star this repository if you find it useful! ⭐

</div>
