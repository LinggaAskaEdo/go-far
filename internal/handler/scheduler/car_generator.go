package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	cfg "go-far/internal/infra/scheduler"
	"go-far/internal/model/dto"
	"go-far/internal/service/car"
	"go-far/internal/service/user"
	"go-far/internal/util"

	"github.com/rs/zerolog"
)

var (
	carColors = []string{
		"Pearl White", "Midnight Black", "Silver Metallic",
		"Ruby Red", "Navy Blue", "Crimson Red",
		"Forest Green", "Ocean Blue", "Sunset Orange",
		"Champagne Gold",
	}

	licenseLetters = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	licenseNumbers = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
)

type CarGeneratorJob struct {
	carService  car.CarServiceItf
	userService user.UserServiceItf
	log         *zerolog.Logger
	config      *cfg.CarGeneratorJobOptions
	httpClient  *http.Client
	nhtsaURL    string
	carCache    []carInfo
	mu          sync.Mutex
	cacheMu     sync.Mutex
}

type carInfo struct {
	Brand  string
	Model  string
	Colors []string
	Year   int
}

type makeInfo struct {
	MakeName string
	MakeID   int
}

type nhtsaMakesResponse struct {
	Result []struct {
		MakeName string `json:"MakeName"`
		MakeID   int    `json:"MakeId"`
	} `json:"Results"`
	Count int `json:"Count"`
}

type nhtsaModelsResponse struct {
	Result []struct {
		ModelName string `json:"Model_Name"`
	} `json:"Results"`
	Count int `json:"Count"`
}

type carData struct {
	Brand        string
	Model        string
	Color        string
	LicensePlate string
	Year         int
	IsAvailable  bool
}

func InitCarGeneratorJob(log *zerolog.Logger, carService car.CarServiceItf, userService user.UserServiceItf, opts *cfg.CarGeneratorJobOptions, httpClient *http.Client, nhtsaURL string) *CarGeneratorJob {
	return &CarGeneratorJob{
		log:         log,
		carService:  carService,
		userService: userService,
		config:      opts,
		httpClient:  httpClient,
		carCache:    make([]carInfo, 0),
		nhtsaURL:    nhtsaURL,
	}
}

func (j *CarGeneratorJob) Name() string {
	return "CarGeneratorJob"
}

func (j *CarGeneratorJob) Schedule() string {
	return j.config.Cron
}

func (j *CarGeneratorJob) Run(ctx context.Context) error {
	if !j.config.Enabled {
		j.log.Debug().Msg("CarGeneratorJob is disabled")
		return nil
	}

	j.fetchCarDataFromAPI(ctx)

	j.log.Info().
		Int("batch_size", j.config.BatchSize).
		Msg("Generating random cars")

	filter := dto.UserFilter{Page: 1, PageSize: 100}
	cacheControl := dto.CacheControl{}
	users, _, err := j.userService.ListUsers(ctx, cacheControl, &filter)
	if err != nil || users == nil || len(*users) == 0 {
		j.log.Warn().Err(err).Msg("No users found to assign cars to")
		return nil
	}

	userList := *users

	j.mu.Lock()
	defer j.mu.Unlock()

	successCount := 0
	for i := 0; i < j.config.BatchSize; i++ {
		carData := j.generateRandomCar()

		owner := userList[util.RandomInt(len(userList))]

		req := dto.CreateCarRequest{
			Brand:        carData.Brand,
			Model:        carData.Model,
			Year:         carData.Year,
			Color:        carData.Color,
			LicensePlate: carData.LicensePlate,
		}

		_, err := j.carService.CreateCar(ctx, req, owner.ID)
		if err != nil {
			j.log.Warn().
				Err(err).
				Str("license_plate", carData.LicensePlate).
				Msg("Failed to create car")
			continue
		}

		successCount++
		j.log.Debug().
			Str("brand", carData.Brand).
			Str("model", carData.Model).
			Str("license_plate", carData.LicensePlate).
			Msg("Car created successfully")
	}

	j.log.Info().
		Int("success", successCount).
		Int("total", j.config.BatchSize).
		Msg("Car generation batch completed")

	return nil
}

func (j *CarGeneratorJob) generateRandomCar() *carData {
	carInfo := j.randomCar()
	year := j.config.MinYear + util.RandomInt(j.config.MaxYear-j.config.MinYear+1)
	licensePlate := j.generateLicensePlate()

	return &carData{
		Brand:        carInfo.Brand,
		Model:        carInfo.Model,
		Year:         year,
		Color:        carInfo.Colors[util.RandomInt(len(carInfo.Colors))],
		LicensePlate: licensePlate,
		IsAvailable:  util.RandomFloat32() < 0.7,
	}
}

func (j *CarGeneratorJob) randomCar() *carInfo {
	j.cacheMu.Lock()
	defer j.cacheMu.Unlock()

	if len(j.carCache) == 0 {
		j.log.Warn().Msg("car cache is empty, returning fallback")
		return j.getFallbackCar()
	}

	info := j.carCache[util.RandomInt(len(j.carCache))]

	return &info
}

func (j *CarGeneratorJob) fetchCarDataFromAPI(ctx context.Context) {
	if j.nhtsaURL == "" {
		j.log.Warn().Msg("NHTSA API URL not configured, skipping")
		return
	}

	if j.httpClient == nil {
		j.log.Warn().Msg("HTTP client not available, skipping")
		return
	}

	newCars, err := j.fetchMakesFromAPI(ctx)
	if err != nil {
		j.log.Warn().Err(err).Msg("failed to fetch car data from NHTSA API")
		return
	}

	if len(newCars) == 0 {
		j.log.Warn().Msg("no car data fetched from NHTSA API")
		return
	}

	j.cacheMu.Lock()
	j.carCache = newCars
	j.cacheMu.Unlock()

	j.log.Info().Int("count", len(newCars)).Msg("car data fetched from NHTSA API")
}

func (j *CarGeneratorJob) fetchMakesFromAPI(ctx context.Context) ([]carInfo, error) {
	return j.doFetchMakesFromAPI(ctx)
}

func (j *CarGeneratorJob) doFetchMakesFromAPI(ctx context.Context) ([]carInfo, error) {
	url := fmt.Sprintf("%s/GetMakeForManufacturer?format=json", j.nhtsaURL)
	resp, err := j.httpClient.Get(url)
	if err != nil {
		j.log.Warn().Err(err).Msg("failed to fetch makes from NHTSA")
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var makesResp nhtsaMakesResponse
	if err := json.NewDecoder(resp.Body).Decode(&makesResp); err != nil {
		j.log.Warn().Err(err).Msg("failed to decode makes response")
		return nil, err
	}

	if len(makesResp.Result) == 0 {
		return nil, nil
	}

	util.ShuffleSlice(makesResp.Result)

	makes := make([]makeInfo, len(makesResp.Result))
	for i, m := range makesResp.Result {
		makes[i] = makeInfo{MakeID: m.MakeID, MakeName: m.MakeName}
	}

	result := j.fetchModelsForMakes(ctx, makes)
	return result, nil
}

func (j *CarGeneratorJob) fetchModelsForMakes(ctx context.Context, makes []makeInfo) []carInfo {
	numMakes := min(10, len(makes))
	var newCars []carInfo

	for i := 0; i < numMakes; i++ {
		select {
		case <-ctx.Done():
			return newCars
		default:
		}

		mk := makes[i]
		models := j.fetchModelsForMake(mk.MakeName)
		if len(models) == 0 {
			continue
		}

		numModels := min(2, len(models))
		for m := range numModels {
			newCars = append(newCars, carInfo{
				Brand:  mk.MakeName,
				Model:  models[m],
				Year:   j.config.MinYear + util.RandomInt(j.config.MaxYear-j.config.MinYear+1),
				Colors: j.getRandomColors(3),
			})
		}
	}

	return newCars
}

func (j *CarGeneratorJob) fetchModelsForMake(makeName string) []string {
	modelURL := fmt.Sprintf("%s/GetModelsForMake/%s?format=json", j.nhtsaURL, makeName)
	resp, err := j.httpClient.Get(modelURL)
	if err != nil {
		return nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var modelsResp nhtsaModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil
	}

	if len(modelsResp.Result) == 0 {
		return nil
	}

	util.ShuffleSlice(modelsResp.Result)

	models := make([]string, len(modelsResp.Result))
	for i, r := range modelsResp.Result {
		models[i] = r.ModelName
	}
	return models
}

func (j *CarGeneratorJob) getRandomColors(num int) []string {
	shuffled := make([]string, len(carColors))
	copy(shuffled, carColors)
	util.ShuffleSlice(shuffled)

	if num > len(shuffled) {
		num = len(shuffled)
	}

	return shuffled[:num]
}

func (j *CarGeneratorJob) getFallbackCar() *carInfo {
	fallbacks := []carInfo{
		{Brand: "Toyota", Model: "Camry", Year: 2023, Colors: []string{carColors[0], carColors[1], carColors[2]}},
		{Brand: "Honda", Model: "Civic", Year: 2023, Colors: []string{"Rallye Red", "Crystal Black", "Lunar Silver"}},
		{Brand: "Ford", Model: "F-150", Year: 2023, Colors: []string{"Oxford White", "Race Red", "Antimatter Blue"}},
		{Brand: "Tesla", Model: "Model 3", Year: 2023, Colors: []string{carColors[0], "Solid Black", "Midnight Silver"}},
		{Brand: "BMW", Model: "3 Series", Year: 2023, Colors: []string{"Alpine White", "Black Sapphire", "Mineral Gray"}},
	}

	return &fallbacks[util.RandomInt(len(fallbacks))]
}

func (j *CarGeneratorJob) generateLicensePlate() string {
	format := util.RandomInt(4) // returns 0, 1, 2, or 3
	var plate strings.Builder

	switch format {
	case 0:
		plate.WriteString(licenseLetters[util.RandomInt(len(licenseLetters))])
		plate.WriteString(licenseLetters[util.RandomInt(len(licenseLetters))])
		plate.WriteString(licenseLetters[util.RandomInt(len(licenseLetters))])
		plate.WriteString("-")
		plate.WriteString(licenseNumbers[util.RandomInt(len(licenseNumbers))])
		plate.WriteString(licenseNumbers[util.RandomInt(len(licenseNumbers))])
		plate.WriteString(licenseNumbers[util.RandomInt(len(licenseNumbers))])
		plate.WriteString(licenseNumbers[util.RandomInt(len(licenseNumbers))])
	case 1:
		plate.WriteString(licenseLetters[util.RandomInt(len(licenseLetters))])
		plate.WriteString(licenseLetters[util.RandomInt(len(licenseLetters))])
		plate.WriteString(licenseLetters[util.RandomInt(len(licenseLetters))])
		plate.WriteString(licenseLetters[util.RandomInt(len(licenseLetters))])
		plate.WriteString(licenseLetters[util.RandomInt(len(licenseLetters))])
		plate.WriteString(licenseNumbers[util.RandomInt(len(licenseNumbers))])
		plate.WriteString(licenseNumbers[util.RandomInt(len(licenseNumbers))])
		plate.WriteString(licenseNumbers[util.RandomInt(len(licenseNumbers))])
	case 2:
		plate.WriteString(licenseNumbers[util.RandomInt(len(licenseNumbers))])
		plate.WriteString(licenseNumbers[util.RandomInt(len(licenseNumbers))])
		plate.WriteString(licenseNumbers[util.RandomInt(len(licenseNumbers))])
		plate.WriteString(licenseNumbers[util.RandomInt(len(licenseNumbers))])
		plate.WriteString("-")
		plate.WriteString(licenseLetters[util.RandomInt(len(licenseLetters))])
		plate.WriteString(licenseLetters[util.RandomInt(len(licenseLetters))])
		plate.WriteString(licenseLetters[util.RandomInt(len(licenseLetters))])
		plate.WriteString(licenseLetters[util.RandomInt(len(licenseLetters))])
	default:
		plate.WriteString(licenseLetters[util.RandomInt(len(licenseLetters))])
		plate.WriteString(licenseLetters[util.RandomInt(len(licenseLetters))])
		plate.WriteString(licenseLetters[util.RandomInt(len(licenseLetters))])
		plate.WriteString(licenseNumbers[util.RandomInt(len(licenseNumbers))])
		plate.WriteString(licenseNumbers[util.RandomInt(len(licenseNumbers))])
		plate.WriteString(licenseNumbers[util.RandomInt(len(licenseNumbers))])
		plate.WriteString(licenseNumbers[util.RandomInt(len(licenseNumbers))])
		plate.WriteString(licenseNumbers[util.RandomInt(len(licenseNumbers))])
	}

	return fmt.Sprintf("USA-%s", plate.String())
}
