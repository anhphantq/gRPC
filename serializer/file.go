package serializer

import (
	"fmt"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
)

func WriteProtobufToJSON(message proto.Message, filename string) error{
	data, err := ProtobufToJSON(message)
	if err != nil{
		return fmt.Errorf("%w", err)
	}

	err = ioutil.WriteFile(filename, []byte(data), 0644)
	if err != nil{
		return fmt.Errorf("%w", err)
	}

	return nil
}

func WriteProtobufToBinaryFile(message proto.Message, filename string) error{
	data, err := proto.Marshal(message)
	if (err != nil){
		return fmt.Errorf("%w", err)
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil{
		return fmt.Errorf("%w", err)
	}

	return nil
}

func ReadProtobufFromBinaryFile(filename string, message proto.Message) error{
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = proto.Unmarshal(data, message)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}