# LinkedIn Clone API

## 📖 Overview

Postman: https://www.postman.com/tyo-team/workspace/dot/collection/32354585-29d3bd01-8e84-471c-89a6-c24beaf5743c?action=share&creator=32354585

LinkedIn Clone API adalah aplikasi backend yang mensimulasikan fitur-fitur utama LinkedIn menggunakan Go (Golang) dengan framework Gin. Aplikasi ini menyediakan fitur manajemen pengguna, posting, job listings, dan aplikasi kerja dengan sistem autentikasi JWT yang lengkap.

## 🎯 Features

### Core Features
- **User Management**: Registrasi, login, profil pengguna, upload foto profil
- **Posts**: Membuat, membaca, update, delete posts dengan dukungan gambar
- **Jobs**: CRUD job listings dengan aplikasi kerja
- **Applications**: Sistem apply pekerjaan dengan upload resume
- **Authentication**: JWT-based authentication dengan refresh token
- **File Upload**: Upload gambar dan dokumen ke AWS S3
- **Email Service**: Verifikasi email dan reset password

### Technical Features
- **Security**: Rate limiting, CORS, XSS protection, SQL injection protection
- **Logging**: Structured logging dengan tracing
- **Middleware**: Comprehensive middleware stack
- **Validation**: Request validation dengan custom rules
- **Error Handling**: Centralized error handling
- **Database**: PostgreSQL dengan GORM ORM

## 🏗️ Architecture & Design Patterns

### Clean Architecture
Aplikasi ini menggunakan **Clean Architecture** dengan **Domain-Driven Design (DDD)** approach. Berikut alasan penggunaan pattern ini:

#### **Alasan Penggunaan Clean Architecture:**

1. **Separation of Concerns**: Setiap layer memiliki tanggung jawab yang jelas dan terpisah
2. **Dependency Inversion**: Dependency mengalir dari luar ke dalam, business logic tidak bergantung pada framework
3. **Testability**: Mudah untuk unit testing karena dependency dapat di-mock
4. **Maintainability**: Struktur yang jelas memudahkan maintenance dan development
5. **Scalability**: Mudah untuk menambah fitur baru tanpa mengubah existing code

#### **Layer Structure:**

```
├── cmd/app/                    # Application entry point
├── internal/
│   ├── api/                    # API Layer (Controllers/Handlers)
│   │   ├── auth/
│   │   ├── user/
│   │   ├── post/
│   │   └── job/
│   ├── domain/                 # Domain Layer (Business Logic)
│   │   ├── entities/           # Domain entities
│   │   └── repositories/       # Repository interfaces
│   ├── infrastructure/         # Infrastructure Layer
│   │   └── database/
│   ├── config/                 # Configuration
│   └── middleware/             # HTTP middleware
└── pkg/                        # Shared packages
    ├── auth/                   # JWT service
    ├── logger/                 # Logging
    ├── storage/                # File storage
    └── response/               # HTTP responses
```

### **Layering Pattern Benefits:**

1. **API Layer**: Handles HTTP requests/responses, validation, dan serialization
2. **Service Layer**: Business logic dan use cases
3. **Repository Layer**: Data access abstraction
4. **Infrastructure Layer**: External dependencies (database, storage, email)

## 🚀 Tech Stack

- **Framework**: Gin (HTTP web framework)
- **Database**: PostgreSQL
- **ORM**: GORM
- **Authentication**: JWT (JSON Web Tokens)
- **File Storage**: AWS S3
- **Cache**: Redis
- **Email**: SMTP
- **Logging**: Logrus with structured logging
- **Validation**: Go Playground Validator
- **Testing**: Go testing package

## 📋 Prerequisites

- Go 1.19 or higher
- PostgreSQL 12+
- Redis 6+
- AWS S3 Bucket
- SMTP Server (Gmail/SendGrid)

## ⚙️ Installation & Setup

### 1. Clone Repository
```bash
git clone <repository-url>
cd linkedin-clone
```

### 2. Environment Configuration
Create `.env` file:
```env
# Server Configuration
PORT=8080
ENVIRONMENT=development

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=linkedin_clone
DB_SSLMODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration
JWT_SECRET=your_super_secret_jwt_key
JWT_EXPIRY_HOURS=24

# AWS S3 Configuration
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key
AWS_REGION=us-east-1
S3_BUCKET=your_s3_bucket_name

# SMTP Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json
```

### 3. Database Setup
```bash
# Create PostgreSQL database
createdb linkedin_clone

# Run migrations (automatic on startup)
go run cmd/app/main.go
```

### 4. Install Dependencies
```bash
go mod download
```

### 5. Run Application
```bash
go run cmd/app/main.go
```

Application will be available at: `http://localhost:8080`

## 📚 API Documentation

### Base URL
```
http://localhost:8080/api/v1
```

### Health Check Endpoints
```http
GET /health          # Health check
GET /ready           # Readiness check
GET /metrics         # Application metrics
```

### Authentication Endpoints
```http
POST /auth/register           # User registration
POST /auth/login              # User login
POST /auth/verify-email       # Email verification
POST /auth/forgot-password    # Request password reset
POST /auth/reset-password     # Reset password
POST /auth/refresh            # Refresh JWT token
```

### User Endpoints
```http
GET    /users/profile         # Get current user profile
PUT    /users/profile         # Update user profile
POST   /users/profile/picture # Upload profile picture
GET    /users/search          # Search users
GET    /users/:id             # Get user by ID
```

### Post Endpoints
```http
GET    /posts                 # Get user feed
POST   /posts                 # Create new post
GET    /posts/:id             # Get post by ID
PUT    /posts/:id             # Update post
DELETE /posts/:id             # Delete post
POST   /posts/:id/like        # Like post
DELETE /posts/:id/like        # Unlike post
POST   /posts/:id/comments    # Add comment
GET    /posts/:id/comments    # Get comments
GET    /posts/user/:user_id   # Get user posts
```

### Job Endpoints
```http
GET    /jobs                  # Get all jobs
GET    /jobs/search           # Search jobs
GET    /jobs/:id              # Get job by ID
POST   /jobs                  # Create job (auth required)
PUT    /jobs/:id              # Update job (auth required)
DELETE /jobs/:id              # Delete job (auth required)
POST   /jobs/:id/apply        # Apply for job (auth required)
GET    /jobs/:id/applications # Get job applications (auth required)
GET    /jobs/my/jobs          # Get my posted jobs (auth required)
GET    /jobs/my/applications  # Get my applications (auth required)
```

## 🔐 Authentication

### JWT Token Usage
```http
Authorization: Bearer <your_jwt_token>
```

### Registration Flow
1. `POST /auth/register` - Register new user
2. Check email for verification code
3. `POST /auth/verify-email` - Verify email with code
4. User account is now active

### Login Flow
1. `POST /auth/login` - Login with email/password
2. Receive JWT token and user data
3. Use token for authenticated requests

## 🧪 Testing

### Running Tests
```bash
make test-db         
make test-auth  
```

## 📁 Project Structure

```
linkedin-clone/
├── cmd/app/main.go                 # Application entry point
├── internal/
│   ├── api/                        # API handlers and DTOs
│   │   ├── auth/
│   │   │   ├── dto/
│   │   │   ├── handler/
│   │   │   └── service/
│   │   ├── user/
│   │   ├── post/
│   │   └── job/
│   ├── domain/                     # Domain layer
│   │   ├── entities/               # Domain entities
│   │   └── repositories/           # Repository interfaces
│   ├── infrastructure/             # Infrastructure layer
│   │   └── database/
│   ├── config/                     # Configuration management
│   │   ├── server/
│   │   └── routes/
│   └── middleware/                 # HTTP middleware
├── pkg/                           # Shared packages
│   ├── auth/                      # JWT service
│   ├── errors/                    # Error handling
│   ├── logger/                    # Logging utilities
│   ├── redis/                     # Redis client
│   ├── response/                  # HTTP response utilities
│   ├── smtp/                      # Email service
│   ├── storage/                   # File storage (S3)
│   ├── utils/                     # Utility functions
│   └── validator/                 # Request validation
├── tests/                         # Test files
├── docs/                          # Documentation
├── .env.example                   # Environment variables example
├── go.mod                         # Go modules
├── go.sum                         # Go modules checksum
└── README.md                      # This file
```