package types

import "fmt"

type Log struct {
	Message     string `json:"message" bson:"message"`
	ServiceName string `json:"service_name" bson:"service_name"`
}

func (l Log) String() string {
	return fmt.Sprintf("from %s: %s", l.ServiceName, l.Message)
}

type IDatabase interface {
	Save(data Log) error
}
