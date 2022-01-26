package service

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
)

type ImageStore interface {
	Save(laptopID string, imageType string, imageData bytes.Buffer) (string, error)
}

type DiskImageStore struct{
	mutex sync.RWMutex
	imageFolder string
	images map[string]*ImageInfo
}

type ImageInfo struct{
	LaptopID string
	Type string
	Path string
}

func NewDiscImageStore(imageFolder string) *DiskImageStore{
	return &DiskImageStore{
		imageFolder: imageFolder,
		images: make(map[string]*ImageInfo),
	}
}

func (store *DiskImageStore) Save(laptopID string, imageType string, imageData bytes.Buffer) (string, error){
	imageID, err := uuid.NewRandom()

	if (err != nil){
		return "", err
	}

	imagePath := fmt.Sprintf("%s/%s%s", store.imageFolder,imageID, imageType)

	file, err := os.Create(imagePath)
	if err != nil{
		return "", err
	}

	//log.Print(imageData.Len())

	_, err = imageData.WriteTo(file)
	if (err != nil){
		return "", err
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.images[imageID.String()] = &ImageInfo{
		LaptopID: laptopID,
		Type: imageType,
		Path: imagePath,
	}

	return imageID.String(), nil
}