package rpc

import (
	"fmt"
	"log"
	"os"

	"github.com/ghodss/yaml"
	"github.com/only1isus/majorProj/config"
	"github.com/only1isus/majorProj/controller"
	"github.com/only1isus/majorProj/types"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

func getConnectionConfig() (*types.Connection, error) {
	dbConn := types.Database{}
	file, err := config.ReadConfigFile()
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(file, &dbConn); err != nil {
		return nil, err
	}
	return &dbConn.DBConnection, nil
}

func connection() (*grpc.ClientConn, error) {
	connection, err := getConnectionConfig()
	if err != nil {
		return nil, err
	}
	if connection.Host == "" || connection.Port == "" {
		log.Println("Please make sure the database section in the config file is populated")
		os.Exit(1)
	}
	grpcconn, err := grpc.Dial(fmt.Sprintf("%s:%s", connection.Host, connection.Port), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return grpcconn, nil
}

// CommitSensorData takes the data to be sent to the database and a connection.
// If there's an error or the response is false then an error should be returned
func CommitSensorData(data *[]byte) error {
	connection, err := connection()
	if err != nil {
		return err
	}
	defer connection.Close()

	setting, err := getConnectionConfig()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := controller.NewCommitClient(connection)

	resp, err := cc.CommitSensorData(ctx, &controller.SensorData{Data: *data, Key: []byte(setting.Secret)})
	if err != nil || resp.Success == false {
		return fmt.Errorf("something went wrong commiting the data to the database server")
	}
	return nil
}

// CommitLog takes the data to be sent to the database and a connection.
// If there's an error or the response is false then an error should be returned.
func CommitLog(data *[]byte) error {
	connection, err := connection()
	if err != nil {
		return err
	}
	defer connection.Close()

	setting, err := getConnectionConfig()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := controller.NewCommitClient(connection)

	resp, err := cc.CommitLog(ctx, &controller.LogData{Data: *data, Key: []byte(setting.Secret)})
	if err != nil || resp.Success == false {
		return fmt.Errorf("something went wrong commiting the data to the database server")
	}
	return nil
}
