package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	cfg "go-far/internal/infra/scheduler"
	"go-far/internal/model/dto"
	"go-far/internal/model/entity"
	appErr "go-far/internal/model/errors"
	"go-far/internal/service/user"
	"go-far/internal/util"

	"github.com/rs/zerolog"
)

const defaultPassword = "UserPass123!"

var (
	fallbackFirstNames = []string{
		"James", "Mary", "John", "Patricia", "Robert", "Jennifer", "Michael", "Linda",
		"William", "Barbara", "David", "Elizabeth", "Richard", "Susan", "Joseph", "Jessica",
		"Thomas", "Sarah", "Charles", "Karen", "Christopher", "Nancy", "Daniel", "Lisa",
		"Matthew", "Betty", "Anthony", "Margaret", "Mark", "Sandra", "Donald", "Ashley",
		"Steven", "Kimberly", "Paul", "Emily", "Andrew", "Donna", "Joshua", "Michelle",
		"Kenneth", "Dorothy", "Kevin", "Carol", "Brian", "Amanda", "George", "Melissa",
		"Edward", "Deborah", "Ronald", "Stephanie", "Timothy", "Rebecca", "Jason", "Sharon",
	}

	fallbackLastNames = []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
		"Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas",
		"Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson", "White",
		"Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson", "Walker", "Young",
		"Allen", "King", "Wright", "Scott", "Torres", "Nguyen", "Hill", "Flores",
		"Green", "Adams", "Nelson", "Baker", "Hall", "Rivera", "Campbell", "Mitchell",
		"Carter", "Roberts", "Gomez", "Phillips", "Evans", "Turner", "Diaz", "Parker",
	}
)

type UserGeneratorJob struct {
	userService user.UserServiceItf
	log         *zerolog.Logger
	config      *cfg.UserGeneratorJobOptions
	httpClient  *http.Client
	mu          sync.Mutex
}

type randomUserResp struct {
	Results []struct {
		Name struct {
			First string `json:"first"`
			Last  string `json:"last"`
		} `json:"name"`
		Email string `json:"email"`
		Login struct {
			Password string `json:"password"`
		} `json:"login"`
		DOB struct {
			Age int `json:"age"`
		} `json:"dob"`
	} `json:"results"`
}

type randomUser struct {
	Name struct {
		First string
		Last  string
	}
	Email string
	DOB   struct {
		Age int
	}
}

func InitUserGeneratorJob(log *zerolog.Logger, userService user.UserServiceItf, opts *cfg.UserGeneratorJobOptions, httpClient *http.Client) *UserGeneratorJob {
	return &UserGeneratorJob{
		log:         log,
		userService: userService,
		config:      opts,
		httpClient:  httpClient,
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

	users, err := j.fetchRandomUsersFromAPI(ctx, j.config.BatchSize)
	if err != nil {
		j.log.Warn().Err(err).Msg("Failed to fetch from randomuser.me API, using fallback")
		return j.runWithFallback(ctx)
	}

	successCount := 0
	for _, u := range users {
		req := dto.CreateUserRequest{
			Name:     fmt.Sprintf("%s %s", u.Name.First, u.Name.Last),
			Email:    strings.ToLower(u.Email),
			Password: defaultPassword,
			Age:      u.DOB.Age,
			Role:     entity.RoleUser,
		}

		_, err := j.userService.CreateUser(ctx, req)
		if err != nil {
			j.log.Warn().
				Err(err).
				Str("email", u.Email).
				Msg("Failed to create user")
			continue
		}

		successCount++
		j.log.Debug().
			Str("name", req.Name).
			Str("email", u.Email).
			Int("age", u.DOB.Age).
			Msg("User created successfully")
	}

	j.log.Info().
		Int("success", successCount).
		Int("total", j.config.BatchSize).
		Msg("User generation batch completed")

	return nil
}

func (j *UserGeneratorJob) fetchRandomUsersFromAPI(ctx context.Context, count int) ([]randomUser, error) {
	return j.doFetchRandomUsersFromAPI(ctx, count)
}

func (j *UserGeneratorJob) doFetchRandomUsersFromAPI(ctx context.Context, count int) ([]randomUser, error) {
	url := fmt.Sprintf("%s?results=%d&format=json", j.config.RandomUserURL, count)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, appErr.New(fmt.Sprintf("API returned status: %d", resp.StatusCode), appErr.CodeHTTPExternalAPI)

	}

	var apiResp randomUserResp
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if len(apiResp.Results) == 0 {
		return nil, appErr.New("no results from API", appErr.CodeHTTPExternalAPI)
	}

	users := make([]randomUser, len(apiResp.Results))

	for i, r := range apiResp.Results {
		users[i] = randomUser{
			Name:  struct{ First, Last string }{First: r.Name.First, Last: r.Name.Last},
			Email: r.Email,
			DOB:   struct{ Age int }{Age: r.DOB.Age},
		}
	}

	return users, nil
}

func (j *UserGeneratorJob) runWithFallback(ctx context.Context) error {
	j.log.Info().
		Int("batch_size", j.config.BatchSize).
		Msg("Generating random users (fallback)")

	j.mu.Lock()
	defer j.mu.Unlock()

	successCount := 0
	for range j.config.BatchSize {
		firstName := fallbackFirstNames[util.RandomInt(len(fallbackFirstNames))]
		lastName := fallbackLastNames[util.RandomInt(len(fallbackLastNames))]
		name := fmt.Sprintf("%s %s", firstName, lastName)

		timestamp := time.Now().Unix()
		email := fmt.Sprintf("%s.%s.%d@gofar.com", firstName, lastName, timestamp+int64(util.RandomInt(1000)))
		age := j.config.MinAge + util.RandomInt(j.config.MaxAge-j.config.MinAge+1)

		req := dto.CreateUserRequest{
			Name:     name,
			Email:    strings.ToLower(email),
			Password: defaultPassword,
			Age:      age,
			Role:     entity.RoleUser,
		}

		_, err := j.userService.CreateUser(ctx, req)
		if err != nil {
			j.log.Warn().
				Err(err).
				Str("email", email).
				Msg("Failed to create user")
			continue
		}

		successCount++
		j.log.Debug().
			Str("name", name).
			Str("email", email).
			Int("age", age).
			Msg("User created successfully")
	}

	j.log.Info().
		Int("success", successCount).
		Int("total", j.config.BatchSize).
		Msg("User generation batch completed (fallback)")

	return nil
}
