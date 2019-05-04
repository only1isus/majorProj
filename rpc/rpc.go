package rpc

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/ghodss/yaml"
	"github.com/only1isus/majorProj/config"
	"github.com/only1isus/majorProj/controller"
	"github.com/only1isus/majorProj/notification"
	db "github.com/only1isus/majorProj/server/database"
	"github.com/only1isus/majorProj/types"
	"github.com/segmentio/ksuid"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

type NotificationSVR struct{}
type CommitSVR struct{}

func (s *CommitSVR) CommitSensorData(ctx context.Context, data *controller.SensorData) (*controller.SuccessResponse, error) {
	d := types.SensorEntry{}
	err := json.Unmarshal(data.Data, &d)
	if err != nil {
		return &controller.SuccessResponse{Success: false}, err
	}
	err = db.AddSensorEntry(data.Key, []byte(ksuid.New().String()), d)
	if err != nil {
		return &controller.SuccessResponse{Success: false}, err
	}
	return &controller.SuccessResponse{Success: true}, nil
}

func (s *CommitSVR) CommitLog(ctx context.Context, data *controller.LogData) (*controller.SuccessResponse, error) {
	l := types.LogEntry{}
	if err := json.Unmarshal(data.Data, &l); err != nil {
		return &controller.SuccessResponse{Success: false}, err
	}
	err := db.AddLogEntry(data.Key, []byte(ksuid.New().String()), l)
	if err != nil {
		return &controller.SuccessResponse{Success: false}, err
	}
	return &controller.SuccessResponse{Success: true}, nil
}

func getDBConnectionConfig() (*types.DBConnection, error) {
	dbConn := types.Database{}
	file, err := config.ReadConfigFile()
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(file, &dbConn); err != nil {
		return nil, err
	}

	if dbConn.Connection.Host == "" || dbConn.Connection.Port == "" {
		return nil, fmt.Errorf("no host or port name")
	}
	return &dbConn.Connection, nil
}

func getNotificationConnectionConfig() (*types.NotificationConnection, error) {
	noti := types.Notification{}
	file, err := config.ReadConfigFile()
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(file, &noti); err != nil {
		return nil, err
	}
	if noti.Connection.Host == "" || noti.Connection.Port == "" {
		return nil, fmt.Errorf("no host or port name")
	}
	return &noti.Connection, nil
}

func dbconnection() (*grpc.ClientConn, error) {
	connection, err := getDBConnectionConfig()
	if err != nil {
		return nil, err
	}
	grpcconn, err := grpc.Dial(fmt.Sprintf("%s:%s", connection.Host, connection.Port), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return grpcconn, nil
}

func notificationconnection() (*grpc.ClientConn, error) {
	connection, err := getNotificationConnectionConfig()
	if err != nil {
		return nil, err
	}
	grpcconn, err := grpc.Dial(fmt.Sprintf("%s:%s", connection.Host, connection.Port), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return grpcconn, nil
}

func NewServer(grpcsrv *grpc.Server, grpcPort string) {
	conn, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Printf("cannot create a connection on port %s\n", grpcPort)
		os.Exit(1)
	}
	log.Printf("gRPC server running on port %v\n", grpcPort)
	if err := grpcsrv.Serve(conn); err != nil {
		log.Fatalf("failed to create gRPC serve: %v\n", err)
		os.Exit(1)
	}
}

// CommitSensorData takes the data to be sent to the database and a connection.
// If there's an error or the response is false then an error should be returned
func CommitSensorData(data *[]byte) error {
	connection, err := dbconnection()
	if err != nil {
		return err
	}
	defer connection.Close()

	setting, err := getDBConnectionConfig()
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
	connection, err := dbconnection()
	if err != nil {
		return err
	}
	defer connection.Close()

	setting, err := getDBConnectionConfig()
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

func SendNotification(message string, reciever string) error {
	connection, err := notificationconnection()
	if err != nil {
		return err
	}
	defer connection.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nc := notification.NewNotifyClient(connection)
	resp, err := nc.Send(ctx, &notification.Params{Msg: message, Reciever: reciever})
	if err != nil || resp.Success == false {
		return err
	}
	return nil
}
