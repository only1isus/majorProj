package consts

type BucketName string
type OutputDevice string
type BucketFilter string
type AnalogSensor string

type I2CSensor string

type ADS1115Device string

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
	PHSensor         AnalogSensor = "ph"

	Sth3xTemperature I2CSensor = "sth3xtemperature"
	Sth3xHumidity    I2CSensor = "sth3xhumidity"

	ADS1115Device1 ADS1115Device = "ads1115_1"
	ADS1115Device2 ADS1115Device = "ads1115_2"
)
