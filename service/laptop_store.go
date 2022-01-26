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
	Search (filter *pb.Filter, found func(*pb.Laptop) error) error
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
	other, err := deepCopy(laptop)
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

	return deepCopy(laptop)
}

func (store *InMemoryLaptopStore) Search(filter *pb.Filter, found func(*pb.Laptop) error) error{
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	for _, laptop := range store.data{
		if (isQualified(filter, laptop)){
			other, err := deepCopy(laptop)
			if err != nil{
				return err
			}

			err = found(other)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() {
		return false
	}

	if laptop.GetCpu().GetNumberCores() < filter.GetMinCpuCores() {
		return false
	}

	if laptop.GetCpu().GetMinGhz() < filter.GetMinCpuGhz() {
		return false
	}

	if toBit(laptop.GetRam()) < toBit(filter.GetMinRam()) {
		return false
	}

	return true
}

func toBit(memory *pb.Memory) uint64 {
	value := memory.GetValue()

	switch memory.GetUnit() {
	case pb.Memory_BIT:
		return value
	case pb.Memory_BYTE:
		return value << 3 // 8 = 2^3
	case pb.Memory_KILOBYTE:
		return value << 13 // 1024 * 8 = 2^10 * 2^3 = 2^13
	case pb.Memory_MEGABYTE:
		return value << 23
	case pb.Memory_GIGABYTE:
		return value << 33
	case pb.Memory_TERABYTE:
		return value << 43
	default:
		return 0
	}
}

func deepCopy(laptop *pb.Laptop) (*pb.Laptop, error) {
	other := &pb.Laptop{}

	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data: %w", err)
	}

	return other, nil
}