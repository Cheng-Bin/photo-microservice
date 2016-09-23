/**
*
*  存取、设置任务微服务
*
**/
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

// Task struct
type Task struct {
	ID    int `json:"id"`
	State int `json:"state"`
}

var datastore []Task
var datastoreMutex sync.RWMutex
var oldestNotFinishedTask int
var oNFTMutext sync.RWMutex

func main() {

	if !registerInKVStore() {
		return
	}

	datastore = make([]Task, 0)
	oldestNotFinishedTask = 0
	datastoreMutex = sync.RWMutex{}
	oNFTMutext = sync.RWMutex{}

	http.HandleFunc("/getById", getByID)
	http.HandleFunc("/newTask", newTask)
	http.HandleFunc("/getNewTask", getNewTask)
	http.HandleFunc("/finishedTask", finishedTask)
	http.HandleFunc("/setById", setByID)
	http.HandleFunc("/list", list)
	http.ListenAndServe(":3001", nil)

}

func registerInKVStore() bool {
	if len(os.Args) < 3 {
		fmt.Println("Error: Task few arguments.")
		return false
	}

	databaseAddress := os.Args[0]
	keyValueStoreAddress := os.Args[1]

	response, err := http.Post("http://"+keyValueStoreAddress+"/set?key=databaseAddress&value="+databaseAddress, "", nil)
	if err != nil {
		fmt.Println(err)
		return false
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if response.StatusCode != http.StatusOK {
		fmt.Println("Error: Failure whene contacting key-value store: ", string(data))
		return false
	}

	return true
}

// 通过id获取任务
func getByID(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		if len(values.Get("id")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Wrong input")
			return
		}
		id, err := strconv.Atoi(string(values.Get("id")))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err)
			return
		}
		datastoreMutex.RLock()
		bIsInError := err != nil || id >= len(datastore)
		datastoreMutex.RUnlock()

		if bIsInError {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Wrong input")
			return
		}
		datastoreMutex.RLocker()
		value := datastore[id]
		datastoreMutex.RUnlock()

		response, err := json.Marshal(value)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err)
			return
		}
		fmt.Fprint(w, string(response))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only Get accepted")
	}
}

// 新建任务
func newTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		datastoreMutex.Lock()
		taskToAdd := Task{
			ID:    len(datastore),
			State: 0,
		}
		datastore[taskToAdd.ID] = taskToAdd
		datastoreMutex.RUnlock()
		fmt.Fprint(w, http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted")
	}
}

// 获取新任务
func getNewTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		bErrored := false

		datastoreMutex.RLock()
		if len(datastore) == 0 {
			bErrored = true
		}
		datastoreMutex.RUnlock()

		if bErrored {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: No non-started task.")
		}

		taskToSend := Task{ID: -1, State: 0}

		oNFTMutext.Lock()
		datastoreMutex.Lock()
		for i := oldestNotFinishedTask; i < len(datastore); i++ {
			if datastore[i].State == 2 && i == oldestNotFinishedTask {
				oldestNotFinishedTask++
				continue
			}
			if datastore[i].State == 0 {
				datastore[i] = Task{ID: i, State: 1}
				taskToSend = datastore[i]
				break
			}
		}
		datastoreMutex.Unlock()
		oNFTMutext.Unlock()

		if taskToSend.ID == -1 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: No non-started task.")
			return
		}

		myID := taskToSend.ID

		go func() {
			time.Sleep(time.Second * 120)
			datastoreMutex.Lock()
			if datastore[myID].State == 1 {
				datastore[myID] = Task{ID: myID, State: 0}
			}
			datastoreMutex.Unlock()
		}()

		response, err := json.Marshal(taskToSend)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err)
			return
		}

		fmt.Fprint(w, string(response))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted.")
	}
}

// 完成任务
func finishedTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		if len(values.Get("id")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Wrong input")
			return
		}

		id, err := strconv.Atoi(string(values.Get("id")))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err)
			return
		}

		updateTask := Task{ID: id, State: 2}
		bErrored := false

		datastoreMutex.Lock()
		if datastore[id].State == 1 {
			datastore[id] = updateTask
		} else {
			bErrored = true
		}
		datastoreMutex.RUnlock()

		if bErrored {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: Wrong input")
			return
		}

		fmt.Fprint(w, "success")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted.")
	}
}

//  设置任务
func setByID(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		taskToSet := Task{}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err)
			return
		}
		err = json.Unmarshal([]byte(data), &taskToSet)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err)
			return
		}
		bErrored := false
		datastoreMutex.Lock()
		if taskToSet.ID >= len(datastore) || taskToSet.State > 2 || taskToSet.State < 0 {
			bErrored = true
		} else {
			datastore[taskToSet.ID] = taskToSet
		}
		datastoreMutex.RUnlock()

		if bErrored {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: Wrong input")
			return
		}
		fmt.Fprint(w, "success")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted.")
	}
}

//  获取任务列表
func list(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		datastoreMutex.RLock()
		for key, value := range datastore {
			fmt.Fprintln(w, key, ": ", "id: ", value.ID, " state: ", value.State)
		}
		datastoreMutex.RUnlock()
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only get accepted.")
	}
}
