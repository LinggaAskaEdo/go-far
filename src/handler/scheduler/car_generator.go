package scheduler

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	cfg "go-far/src/config/scheduler"
	"go-far/src/model/dto"
	"go-far/src/service/car"
	"go-far/src/service/user"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type CarGeneratorJob struct {
	log         zerolog.Logger
	carService  car.CarServiceItf
	userService user.UserServiceItf
	config      cfg.CarGeneratorJobOptions
	rng         *rand.Rand
	mu          sync.Mutex
}

func InitCarGeneratorJob(log zerolog.Logger, carService car.CarServiceItf, userService user.UserServiceItf, cfg cfg.CarGeneratorJobOptions) *CarGeneratorJob {
	return &CarGeneratorJob{
		log:         log,
		carService:  carService,
		userService: userService,
		config:      cfg,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
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

	j.log.Info().
		Int("batch_size", j.config.BatchSize).
		Msg("Generating random cars")

	filter := dto.UserFilter{Page: 1, PageSize: 100}
	cacheControl := dto.CacheControl{}
	users, _, err := j.userService.ListUsers(ctx, cacheControl, filter)
	if err != nil || users == nil || len(*users) == 0 {
		j.log.Warn().Err(err).Msg("No users found to assign cars to")
		return nil
	}

	userList := *users

	successCount := 0
	for i := 0; i < j.config.BatchSize; i++ {
		carData := j.generateRandomCar()

		owner := userList[j.rng.Intn(len(userList))]
		ownerID, _ := uuid.Parse(owner.ID)

		req := dto.CreateCarRequest{
			UserID:       ownerID,
			Brand:        carData.Brand,
			Model:        carData.Model,
			Year:         carData.Year,
			Color:        carData.Color,
			LicensePlate: carData.LicensePlate,
		}

		_, err := j.carService.CreateCar(ctx, req)
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

type carData struct {
	Brand        string
	Model        string
	Year         int
	Color        string
	LicensePlate string
	IsAvailable  bool
}

func (j *CarGeneratorJob) generateRandomCar() *carData {
	j.mu.Lock()
	defer j.mu.Unlock()

	carInfo := j.randomCar()
	year := j.config.MinYear + j.rng.Intn(j.config.MaxYear-j.config.MinYear+1)
	licensePlate := j.generateLicensePlate()

	return &carData{
		Brand:        carInfo.Brand,
		Model:        carInfo.Model,
		Year:         year,
		Color:        carInfo.Colors[j.rng.Intn(len(carInfo.Colors))],
		LicensePlate: licensePlate,
		IsAvailable:  j.rng.Float32() < 0.7,
	}
}

type carInfo struct {
	Brand  string
	Model  string
	Colors []string
}

func (j *CarGeneratorJob) randomCar() *carInfo {
	cars := []carInfo{
		{Brand: "Toyota", Model: "Camry", Colors: []string{"Pearl White", "Midnight Black", "Silver Metallic", "Ruby Red", "Navy Blue"}},
		{Brand: "Toyota", Model: "Corolla", Colors: []string{"Super White", "Black", "Classic Silver", "Barcelona Red", "Hydro Blue"}},
		{Brand: "Toyota", Model: "RAV4", Colors: []string{"Magnetic Gray", "Black", "Lunar Rock", "Blue Fusion", "Ruby Flare"}},
		{Brand: "Toyota", Model: "Highlander", Colors: []string{"Blizzard Pearl", "Magnetic Gray", "Attitude Black", "Blueprint", "Ruby Flare"}},
		{Brand: "Honda", Model: "Civic", Colors: []string{"Rallye Red", "Crystal Black", "Lunar Silver", "Modern Steel", "Aegean Blue"}},
		{Brand: "Honda", Model: "Accord", Colors: []string{"Platinum White", "Crystal Black", "Lunar Silver", "Modern Steel", "Still Night"}},
		{Brand: "Honda", Model: "CR-V", Colors: []string{"Sonic Gray", "Crystal Black", "Lunar Silver", "Obsidian Blue", "Radiant Red"}},
		{Brand: "Honda", Model: "Pilot", Colors: []string{"Platinum White", "Crystal Black", "Lunar Silver", "Steel Sapphire", "Obsidian Blue"}},
		{Brand: "Ford", Model: "F-150", Colors: []string{"Oxford White", "Race Red", "Antimatter Blue", "Carbonized Gray", "Agate Black"}},
		{Brand: "Ford", Model: "Mustang", Colors: []string{"Race Red", "Oxford White", "Twister Orange", "Velocity Blue", "Dark Highland Green"}},
		{Brand: "Ford", Model: "Explorer", Colors: []string{"Rapid Red", "Agate Black", "Iconic Silver", "Blue Metallic", "Stone Gray"}},
		{Brand: "Ford", Model: "Escape", Colors: []string{"Rapid Red", "Agate Black", "Iconic Silver", "Sedona", "Baltic Sea"}},
		{Brand: "Chevrolet", Model: "Silverado", Colors: []string{"Summit White", "Black", "Red Hot", "Silver Ice", "Oakwood"}},
		{Brand: "Chevrolet", Model: "Malibu", Colors: []string{"Summit White", "Black", "Mosaic Gray", "Cherry Bomb", "Cayenne Orange"}},
		{Brand: "Chevrolet", Model: "Equinox", Colors: []string{"Summit White", "Black", "Mosaic Gray", "Cayenne", "Kinetic Blue"}},
		{Brand: "Chevrolet", Model: "Tahoe", Colors: []string{"Summit White", "Black", "Cherry Bomb", "Iridescent Pearl", "Satin Steel"}},
		{Brand: "BMW", Model: "3 Series", Colors: []string{"Alpine White", "Black Sapphire", "Mineral Gray", "Jet Black", "Melbourne Red"}},
		{Brand: "BMW", Model: "5 Series", Colors: []string{"Alpine White", "Black Sapphire", "Mineral Gray", "Phytonic Blue", "Tanzanite Blue"}},
		{Brand: "BMW", Model: "X3", Colors: []string{"Alpine White", "Black Sapphire", "Mineral Gray", "Phytonic Blue", "Brooklyn Gray"}},
		{Brand: "BMW", Model: "X5", Colors: []string{"Alpine White", "Black Sapphire", "Mineral Gray", "Phytonic Blue", "Tanzanite Blue"}},
		{Brand: "Mercedes-Benz", Model: "C-Class", Colors: []string{"Polar White", "Obsidian Black", "Selenite Gray", "Selenite Pearl", "Hyacinth Red"}},
		{Brand: "Mercedes-Benz", Model: "E-Class", Colors: []string{"Polar White", "Obsidian Black", "Selenite Gray", "Designo Diamond", "Hyacinth Red"}},
		{Brand: "Mercedes-Benz", Model: "GLC", Colors: []string{"Polar White", "Obsidian Black", "Selenite Gray", "Mojave", "Denim Blue"}},
		{Brand: "Mercedes-Benz", Model: "GLE", Colors: []string{"Polar White", "Obsidian Black", "Selenite Gray", "Mojave", "Selenite Pearl"}},
		{Brand: "Audi", Model: "A4", Colors: []string{"Glacier White", "Mythos Black", "Navarra Blue", "Daytona Gray", "Progressive Red"}},
		{Brand: "Audi", Model: "A6", Colors: []string{"Glacier White", "Mythos Black", "Navarra Blue", "Firmament Blue", "Seville Red"}},
		{Brand: "Audi", Model: "Q5", Colors: []string{"Glacier White", "Mythos Black", "Navarra Blue", "Daytona Gray", "Progressive Red"}},
		{Brand: "Audi", Model: "Q7", Colors: []string{"Glacier White", "Mythos Black", "Orca Black", "Night Suite", "Barolo Brown"}},
		{Brand: "Tesla", Model: "Model 3", Colors: []string{"Pearl White", "Solid Black", "Midnight Silver", "Deep Blue", "Red Multi-Coat"}},
		{Brand: "Tesla", Model: "Model Y", Colors: []string{"Pearl White", "Solid Black", "Midnight Silver", "Deep Blue", "Red Multi-Coat"}},
		{Brand: "Tesla", Model: "Model S", Colors: []string{"Pearl White", "Solid Black", "Midnight Silver", "Ultra Red", "Deep Blue"}},
		{Brand: "Tesla", Model: "Model X", Colors: []string{"Pearl White", "Solid Black", "Midnight Silver", "Ultra Red", "Deep Blue"}},
		{Brand: "Nissan", Model: "Altima", Colors: []string{"Super Black", "Gun Metallic", "Scarlet Ember", "Pearl White", "Deep Blue Pearl"}},
		{Brand: "Nissan", Model: "Sentra", Colors: []string{"Super Black", "Gun Metallic", "Electric Blue", "Pearl White", "Monarch Orange"}},
		{Brand: "Nissan", Model: "Rogue", Colors: []string{"Super Black", "Gun Metallic", "Pearl White", "Caspian Blue", "Gun Metallic"}},
		{Brand: "Hyundai", Model: "Sonata", Colors: []string{"Phantom Black", "Portofino Gray", "Calypso Red", "Oxford White", "Hyper White"}},
		{Brand: "Hyundai", Model: "Elantra", Colors: []string{"Phantom Black", "Intense Blue", "Fluid Metal", "Cherry Bomb", "Atlas White"}},
		{Brand: "Hyundai", Model: "Tucson", Colors: []string{"Phantom Black", "Amazon Gray", "Intense Blue", "Shimmering Silver", "Sage Gray"}},
		{Brand: "Hyundai", Model: "Santa Fe", Colors: []string{"Phantom Black", "Twilight Black", "Waterfall White", "Taiga Brown", "Sage Gray"}},
		{Brand: "Kia", Model: "Optima", Colors: []string{"Snow White Pearl", "Aurora Black", "Sparkling Silver", "Runway Red", "Pacific Blue"}},
		{Brand: "Kia", Model: "Sportage", Colors: []string{"Snow White Pearl", "Aurora Black", "Steel Gray", "Cherry Black", "Pacific Blue"}},
		{Brand: "Kia", Model: "Telluride", Colors: []string{"Snow White Pearl", "Aurora Black", "Steel Gray", "Wolf Gray", "Everlasting Silver"}},
		{Brand: "Mazda", Model: "Mazda3", Colors: []string{"Soul Red Crystal", "Machine Gray", "Snowflake White", "Jet Black", "Deep Crystal Blue"}},
		{Brand: "Mazda", Model: "Mazda6", Colors: []string{"Soul Red Crystal", "Machine Gray", "Snowflake White", "Jet Black", "Deep Crystal Blue"}},
		{Brand: "Mazda", Model: "CX-5", Colors: []string{"Soul Red Crystal", "Machine Gray", "Snowflake White", "Jet Black", "Deep Crystal Blue"}},
		{Brand: "Mazda", Model: "CX-9", Colors: []string{"Soul Red Crystal", "Machine Gray", "Snowflake White", "Jet Black", "Deep Crystal Blue"}},
		{Brand: "Subaru", Model: "Outback", Colors: []string{"Crystal Black Silica", "Crystal White Pearl", "Magnetite Gray", "Crimson Red Pearl", "Autumn Green"}},
		{Brand: "Subaru", Model: "Forester", Colors: []string{"Crystal Black Silica", "Crystal White Pearl", "Magnetite Gray", "Sepia", "Harbor Blue"}},
		{Brand: "Subaru", Model: "Impreza", Colors: []string{"Crystal Black Silica", "Crystal White Pearl", "Magnetite Gray", "Lapis Blue", "Ocean Blue"}},
		{Brand: "Volkswagen", Model: "Jetta", Colors: []string{"Pure White", "Deep Black", "Tornado Red", "Platinum Gray", "Pyrite Silver"}},
		{Brand: "Volkswagen", Model: "Passat", Colors: []string{"Pure White", "Deep Black", "Tornado Red", "Platinum Gray", "Tourmaline Blue"}},
		{Brand: "Volkswagen", Model: "Tiguan", Colors: []string{"Pure White", "Deep Black", "Tornado Red", "Platinum Gray", "Night Blue"}},
		{Brand: "Volkswagen", Model: "Atlas", Colors: []string{"Pure White", "Deep Black", "Tornado Red", "Platinum Gray", "Fortana Red"}},
		{Brand: "Lexus", Model: "ES", Colors: []string{"Caviar", "Eminent White", "Atomic Silver", "Matte Nocturnal", "Crimson"}},
		{Brand: "Lexus", Model: "RX", Colors: []string{"Caviar", "Eminent White", "Atomic Silver", "Matte Black", "Crimson"}},
		{Brand: "Lexus", Model: "NX", Colors: []string{"Caviar", "Eminent White", "Atomic Silver", "Cobalt", "Cadmium Orange"}},
		{Brand: "Lexus", Model: "IS", Colors: []string{"Caviar", "Eminent White", "Atomic Silver", "Infrared", "Ultrasonic Blue"}},
		{Brand: "Porsche", Model: "911", Colors: []string{"GT Silver", "Jet Black", "Guards Red", "Carrara White", "Racing Yellow"}},
		{Brand: "Porsche", Model: "Cayenne", Colors: []string{"Jet Black", "Carrara White", "Volcano Gray", "Quartzite", "Black Taupe"}},
		{Brand: "Porsche", Model: "Macan", Colors: []string{"Jet Black", "Carrara White", "Volcano Gray", "Dolomite Silver", "Mamba Green"}},
		{Brand: "Jeep", Model: "Wrangler", Colors: []string{"Bright White", "Black", "Firecracker Red", "Sarge Green", "Hydro Blue"}},
		{Brand: "Jeep", Model: "Grand Cherokee", Colors: []string{"Bright White", "Black", "Velvet Red", "Silver Zynith", "Midnight Sky"}},
		{Brand: "Jeep", Model: "Cherokee", Colors: []string{"Bright White", "Black", "Light Brownstone", "Sting Gray", "Hydro Blue"}},
		{Brand: "GMC", Model: "Sierra", Colors: []string{"Summit White", "Onyx Black", "Cardinal Red", "Quicksilver", "Mountain Shadow"}},
		{Brand: "GMC", Model: "Yukon", Colors: []string{"Summit White", "Onyx Black", "Dark Sky", "Quicksilver", "Cayenne"}},
		{Brand: "Ram", Model: "1500", Colors: []string{"Bright White", "Diamond Black", "Flame Red", "Patriot Blue", "Hydro Blue"}},
		{Brand: "Ram", Model: "2500", Colors: []string{"Bright White", "Diamond Black", "Flame Red", "Granite", "Maximum Steel"}},
	}

	return &cars[j.rng.Intn(len(cars))]
}

func (j *CarGeneratorJob) generateLicensePlate() string {
	j.mu.Lock()
	defer j.mu.Unlock()

	letters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	numbers := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

	format := j.rng.Intn(4)
	var plate strings.Builder

	switch format {
	case 0:
		plate.WriteString(letters[j.rng.Intn(len(letters))])
		plate.WriteString(letters[j.rng.Intn(len(letters))])
		plate.WriteString(letters[j.rng.Intn(len(letters))])
		plate.WriteString("-")
		plate.WriteString(numbers[j.rng.Intn(len(numbers))])
		plate.WriteString(numbers[j.rng.Intn(len(numbers))])
		plate.WriteString(numbers[j.rng.Intn(len(numbers))])
		plate.WriteString(numbers[j.rng.Intn(len(numbers))])
	case 1:
		plate.WriteString(letters[j.rng.Intn(len(letters))])
		plate.WriteString(letters[j.rng.Intn(len(letters))])
		plate.WriteString(letters[j.rng.Intn(len(letters))])
		plate.WriteString(letters[j.rng.Intn(len(letters))])
		plate.WriteString(letters[j.rng.Intn(len(letters))])
		plate.WriteString(numbers[j.rng.Intn(len(numbers))])
		plate.WriteString(numbers[j.rng.Intn(len(numbers))])
		plate.WriteString(numbers[j.rng.Intn(len(numbers))])
	case 2:
		plate.WriteString(numbers[j.rng.Intn(len(numbers))])
		plate.WriteString(numbers[j.rng.Intn(len(numbers))])
		plate.WriteString(numbers[j.rng.Intn(len(numbers))])
		plate.WriteString(numbers[j.rng.Intn(len(numbers))])
		plate.WriteString("-")
		plate.WriteString(letters[j.rng.Intn(len(letters))])
		plate.WriteString(letters[j.rng.Intn(len(letters))])
		plate.WriteString(letters[j.rng.Intn(len(letters))])
		plate.WriteString(letters[j.rng.Intn(len(letters))])
	default:
		plate.WriteString(letters[j.rng.Intn(len(letters))])
		plate.WriteString(letters[j.rng.Intn(len(letters))])
		plate.WriteString(letters[j.rng.Intn(len(letters))])
		plate.WriteString(numbers[j.rng.Intn(len(numbers))])
		plate.WriteString(numbers[j.rng.Intn(len(numbers))])
		plate.WriteString(numbers[j.rng.Intn(len(numbers))])
		plate.WriteString(numbers[j.rng.Intn(len(numbers))])
		plate.WriteString(numbers[j.rng.Intn(len(numbers))])
	}

	return fmt.Sprintf("USA-%s", plate.String())
}
