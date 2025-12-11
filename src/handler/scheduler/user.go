package scheduler

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go-far/src/config"
	"go-far/src/domain"
	"go-far/src/service/user"

	"github.com/rs/zerolog"
)

type UserGeneratorJob struct {
	log         zerolog.Logger
	userService user.UserServiceItf
	config      config.UserGeneratorJobOptions
	rng         *rand.Rand
}

func InitUserJobGenerator(log zerolog.Logger, userService user.UserServiceItf, cfg config.UserGeneratorJobOptions) *UserGeneratorJob {
	return &UserGeneratorJob{
		log:         log,
		userService: userService,
		config:      cfg,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (j *UserGeneratorJob) Name() string {
	return "UserGeneratorJob"
}

func (j *UserGeneratorJob) Schedule() string {
	return j.config.Cron
}

func (j *UserGeneratorJob) Run(ctx context.Context) error {
	if !j.config.Enabled {
		j.log.Debug().Msg("UserGeneratorJob is disabled")
		return nil
	}

	j.log.Info().
		Int("batch_size", j.config.BatchSize).
		Msg("Generating random users")

	successCount := 0
	for i := 0; i < j.config.BatchSize; i++ {
		user := j.generateRandomUser()

		req := domain.CreateUserRequest{
			Name:  user.Name,
			Email: user.Email,
			Age:   user.Age,
		}

		_, err := j.userService.CreateUser(ctx, req)
		if err != nil {
			j.log.Warn().
				Err(err).
				Str("email", user.Email).
				Msg("Failed to create user")
			continue
		}

		successCount++
		j.log.Debug().
			Str("name", user.Name).
			Str("email", user.Email).
			Int("age", user.Age).
			Msg("User created successfully")
	}

	j.log.Info().
		Int("success", successCount).
		Int("total", j.config.BatchSize).
		Msg("User generation batch completed")

	return nil
}

func (j *UserGeneratorJob) generateRandomUser() *domain.User {
	firstName := j.randomFirstName()
	lastName := j.randomLastName()
	name := fmt.Sprintf("%s %s", firstName, lastName)

	timestamp := time.Now().Unix()
	email := fmt.Sprintf("%s.%s.%d@example.com",
		firstName,
		lastName,
		timestamp+int64(j.rng.Intn(1000)))

	age := j.config.MinAge + j.rng.Intn(j.config.MaxAge-j.config.MinAge+1)

	return &domain.User{
		Name:  name,
		Email: email,
		Age:   age,
	}
}

func (j *UserGeneratorJob) randomFirstName() string {
	firstNames := []string{
		"James", "Mary", "John", "Patricia", "Robert", "Jennifer", "Michael", "Linda",
		"William", "Barbara", "David", "Elizabeth", "Richard", "Susan", "Joseph", "Jessica",
		"Thomas", "Sarah", "Charles", "Karen", "Christopher", "Nancy", "Daniel", "Lisa",
		"Matthew", "Betty", "Anthony", "Margaret", "Mark", "Sandra", "Donald", "Ashley",
		"Steven", "Kimberly", "Paul", "Emily", "Andrew", "Donna", "Joshua", "Michelle",
		"Kenneth", "Dorothy", "Kevin", "Carol", "Brian", "Amanda", "George", "Melissa",
		"Edward", "Deborah", "Ronald", "Stephanie", "Timothy", "Rebecca", "Jason", "Sharon",
	}

	return firstNames[j.rng.Intn(len(firstNames))]
}

func (j *UserGeneratorJob) randomLastName() string {
	lastNames := []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
		"Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas",
		"Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson", "White",
		"Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson", "Walker", "Young",
		"Allen", "King", "Wright", "Scott", "Torres", "Nguyen", "Hill", "Flores",
		"Green", "Adams", "Nelson", "Baker", "Hall", "Rivera", "Campbell", "Mitchell",
		"Carter", "Roberts", "Gomez", "Phillips", "Evans", "Turner", "Diaz", "Parker",
	}

	return lastNames[j.rng.Intn(len(lastNames))]
}
