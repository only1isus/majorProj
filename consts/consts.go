package consts

type BucketName string
type OutputDevice string
type BucketFilter string

const (
	Temperature BucketFilter = "temperature"
	Humidity    BucketFilter = "humidity"
	PH          BucketFilter = "ph"
	EC          BucketFilter = "ec"
	WaterLevel  BucketFilter = "waterlevel"
	All         BucketFilter = ""

	Sensor BucketName = "sensor"
	User   BucketName = "user"
	Log    BucketName = "log"

	CoolingFan OutputDevice = "coolingFan"
)
