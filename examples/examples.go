package main

import (
	"fmt"
	"log"
	"time"

	"github.com/only1isus/majorProj/consts"
	"github.com/only1isus/majorProj/types"

	"github.com/only1isus/majorProj/control"
)

func recordAndCommit(t *control.Temperature) error {
	temp, err := (*t).Get()
	if err != nil {
		return err
	}

	entry := types.SensorEntry{
		Time:       time.Now().Unix(),
		SensorType: consts.Temperature,
		Value:      *temp,
	}

	err = db.AddEntry(consts.Sensor, []byte(time.Now().Format(time.RFC3339)), entry)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	// // get the tempature
	t := control.Temperature{}
	temp, err := t.Get()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(*temp)

	// // adding the temperature to the database
	// // first step to adding a transaction to the database is to create/prepare the transaction.
	// // This is done by using the Prepare method. Alternatively, the transaction can be created
	// // be manually preparing the SensorTempeture fields.
	entry, err := t.Prepare()
	if err != nil {
		fmt.Println(err)
	}

	// // AddEntry takes the data to be committed along with the key associated with the data and
	// // adds it the the bucket specified. The key must be unique so as to not cause the cause an
	// // error. The key, a bucket within a bucket, can be any unique string. If the key is passed
	// // as time.Now().Format(time.RFC3339) then the data can be filtered when when ready to be
	// // retrieved.
	err = db.AddEntry(consts.Sensor, []byte(time.Now().Format(time.RFC3339)), entry)
	if err != nil {
		fmt.Println(err)
	}

	// the Maintain method takes a target temperature and the fan struct. It constantly probes
	// temperature and turns the cooling fan on when necessary.

	fan, err := control.NewOutputDevice(consts.CoolingFan)
	if err != nil {
		fmt.Println(err)
	}
	if err := t.Maintain(29.6, fan); err != nil {
		fmt.Println(err)
	}

	// collect data every 5 seconds and commit to the database
	temperature := control.Temperature{}

	go func(temp *control.Temperature) {
		for {
			t := time.NewTimer(time.Minute * 5)
			defer t.Stop()
			<-t.C
			if err := recordAndCommit(temp); err != nil {
				log.Printf("got an error while adding to database %v", err.Error())
			}
		}
	}(&temperature)
}
