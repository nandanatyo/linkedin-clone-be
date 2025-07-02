package routes

import (
	"linked-clone/internal/config"
	"linked-clone/pkg/auth"
	"linked-clone/pkg/logger"
	"linked-clone/pkg/redis"
	email "linked-clone/pkg/smtp"
	"linked-clone/pkg/storage"
	validation "linked-clone/pkg/validator"

	"linked-clone/internal/domain/repositories"

	authHandler "linked-clone/internal/api/auth/handler"
	authRepo "linked-clone/internal/api/auth/repository"
	authService "linked-clone/internal/api/auth/service"

	userHandler "linked-clone/internal/api/user/handler"
	userRepo "linked-clone/internal/api/user/repository"
	userService "linked-clone/internal/api/user/service"

	postHandler "linked-clone/internal/api/post/handler"
	postRepo "linked-clone/internal/api/post/repository"
	postService "linked-clone/internal/api/post/service"

	jobHandler "linked-clone/internal/api/job/handler"
	jobRepo "linked-clone/internal/api/job/repository"
	jobService "linked-clone/internal/api/job/service"

	"gorm.io/gorm"
)

type Dependencies struct {
	JWTService     auth.JWTService
	StorageService storage.StorageService
	RedisClient    redis.RedisClient
	EmailService   email.EmailService
	Validator      validation.Validator
	Logger         logger.StructuredLogger

	UserRepository       repositories.UserRepository
	ConnectionRepository repositories.ConnectionRepository
	SessionRepository    repositories.SessionRepository

	AuthHandler       *authHandler.AuthHandler
	UserHandler       *userHandler.UserHandler
	ConnectionHandler *userHandler.ConnectionHandler
	PostHandler       *postHandler.PostHandler
	JobHandler        *jobHandler.JobHandler
}

func InitializeDependencies(cfg *config.Config, db *gorm.DB, logger logger.StructuredLogger) (*Dependencies, error) {

	userRepository := userRepo.NewUserRepository(db)
	connectionRepository := userRepo.NewConnectionRepository(db)
	sessionRepository := authRepo.NewSessionRepository(db)
	postRepository := postRepo.NewPostRepository(db)
	likeRepository := postRepo.NewLikeRepository(db)
	commentRepository := postRepo.NewCommentRepository(db)
	jobRepository := jobRepo.NewJobRepository(db)
	applicationRepository := jobRepo.NewApplicationRepository(db)

	jwtService := auth.NewJWTService(cfg.JWT.SecretKey, cfg.JWT.ExpiryHours, sessionRepository)
	storageService := storage.NewS3StorageService(cfg.AWS.AccessKeyID, cfg.AWS.SecretAccessKey, cfg.AWS.Region, cfg.AWS.S3Bucket)
	redisClient := redis.NewRedisClient(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.DB)
	emailService := email.NewEmailService(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password)
	validator := validation.NewValidator()

	authSvc := authService.NewAuthService(userRepository, jwtService, emailService, redisClient, logger)
	userSvc := userService.NewUserService(userRepository, storageService, logger)
	connectionSvc := userService.NewConnectionService(connectionRepository, userRepository, storageService, logger)
	postSvc := postService.NewPostService(postRepository, userRepository, likeRepository, commentRepository, storageService, logger)
	jobSvc := jobService.NewJobService(jobRepository, applicationRepository, userRepository, storageService, logger)

	authHand := authHandler.NewAuthHandler(authSvc, validator, logger)
	userHand := userHandler.NewUserHandler(userSvc, validator, logger)
	connectionHand := userHandler.NewConnectionHandler(connectionSvc, validator, logger)
	postHand := postHandler.NewPostHandler(postSvc, validator, logger)
	jobHand := jobHandler.NewJobHandler(jobSvc, validator, logger)

	return &Dependencies{

		JWTService:     jwtService,
		StorageService: storageService,
		RedisClient:    redisClient,
		EmailService:   emailService,
		Validator:      validator,
		Logger:         logger,

		UserRepository:       userRepository,
		ConnectionRepository: connectionRepository,
		SessionRepository:    sessionRepository,

		AuthHandler:       authHand,
		UserHandler:       userHand,
		ConnectionHandler: connectionHand,
		PostHandler:       postHand,
		JobHandler:        jobHand,
	}, nil
}
