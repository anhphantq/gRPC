package service

import (
	"errors"
	"fmt"
	"gRPC/pb"
	"sync"

	"github.com/jinzhu/copier"
)

var ErrAlreadyExists = errors.New("record already exists")

type LaptopStore interface {
	Save(laptop *pb.Laptop) error
	Find (ID string) (*pb.Laptop, error)
}

type InMemoryLaptopStore struct{
	mutex sync.RWMutex
	data map[string]*pb.Laptop
}

func NewInMemoryLaptopStore() *InMemoryLaptopStore{
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error{
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if  store.data[laptop.Id] != nil{
		return ErrAlreadyExists
	}

	// deep copy
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if (err != nil){
		return fmt.Errorf("cannot copy data")
	}
	
	store.data[other.Id] = other
	return nil
}

func (store *InMemoryLaptopStore) Find(ID string) (*pb.Laptop, error){
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	laptop := store.data[ID]
	if (laptop == nil){
		return nil, nil
	}

	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil{
		return nil, fmt.Errorf("%w", err)
	}

	return other, nil
}