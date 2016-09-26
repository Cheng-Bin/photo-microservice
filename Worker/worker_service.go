package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Task struct {
	ID    int `json:"id"`
	State int `json:"state"`
}

var masterLocation string
var storageLocation string
var keyValueStoreAddress string

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Error: Too few arguments.")
		return
	}
	keyValueStoreAddress = os.Args[1]
	response, err := http.Get("http//" + keyValueStoreAddress + "/get?key=masterLocation")

	if err != nil || response.StatusCode != http.StatusOK {
		fmt.Println("Error: can't get master address")
		fmt.Println(response.Body)
		return
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	masterLocation = string(data)
	if len(masterLocation) == 0 {
		fmt.Println("Error: can't get master address, Length is zero.")
		return
	}

	response, err = http.Get("http://" + keyValueStoreAddress + "/get?key=storageAddress")
	if err != nil || response.StatusCode != http.StatusOK {
		fmt.Println("Error: can't get storage address.")
		fmt.Println(response.Body)
		return
	}
	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	storageLocation = string(data)
	if len(storageLocation) == 0 {
		fmt.Println("Error: can't get storage address, Length is zero.")
		return
	}

	threadCount, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Error: Counld't parse thread count.")
		return
	}

	myWG := sync.WaitGroup{}
	myWG.Add(threadCount)

	for i := 0; i < threadCount; i++ {
		go func() {
			for {
				myTask, err := getNewTask(masterLocation)
				if err != nil || myTask.ID == -1 {
					fmt.Println(err)
					fmt.Println("Waiting 2 second timeout...")
					time.Sleep(time.Second * 2)
					continue
				}
				myImage, err := getImageFromStorge(storageLocation, myTask)
				if err != nil {
					fmt.Println(err)
					fmt.Println("Waiting 2 second timeout...")
					time.Sleep(time.Second * 2)
					continue
				}

				myImage, err = doWorkOnImage(myImage)
				if err != nil {
					fmt.Println(err)
					fmt.Println("Waiting 2 second timeout...")
					time.Sleep(time.Second * 2)
					registerFinishedTask(masterLocation, myTask)
					continue
				}

				err = sendImageToStorage(storageLocation, myTask, myImage)
				if err != nil {
					fmt.Println(err)
					fmt.Println("Waiting 2 second timeout...")
					time.Sleep(time.Second * 2)
					continue
				}

				err = registerFinishedTask(masterLocation, myTask)
				if err != nil {
					fmt.Println(err)
					fmt.Println("Waiting 2 second timeout...")
					time.Sleep(time.Second * 2)
					continue
				}
			}
		}()
	}

	myWG.Wait()

}

func getNewTask(masterAddress string) (Task, error) {

}

func getImageFromStorge(storageAddress string, myTask Task) (image.Image, error) {

}

func doWorkOnImage(myImage image.Image) (image.Image, error) {

}

func sendImageToStorage(storageAddress string, myTash Task, myImage image.Image) error {

}

func registerFinishedTask(masterAddress string, myTask Task) error {

}
