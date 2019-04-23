package consts

type BucketName string
type OutputDevice string
type BucketFilter string
type AnalogSensor string

const (
	Temperature BucketFilter = "temperature"
	Humidity    BucketFilter = "humidity"
	PH          BucketFilter = "ph"
	EC          BucketFilter = "ec"
	WaterLevel  BucketFilter = "waterlevel"
	All         BucketFilter = ""

	Sensor      BucketName = "sensor"
	User        BucketName = "user"
	Log         BucketName = "log"
	FarmDetails BucketName = "farmdetails"
	Summary     BucketName = "summary"

	CoolingFan      OutputDevice = "coolingFan"
	CirculationPump OutputDevice = "circulationpump"
	GrowLight       OutputDevice = "growlight"

	WaterLevelSensor AnalogSensor = "waterlevel"
)
