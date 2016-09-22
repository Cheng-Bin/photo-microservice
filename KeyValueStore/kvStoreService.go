package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
)

var keyValueStore map[string]string
var kvStoreMutex sync.RWMutex

func main() {

	keyValueStore = make(map[string]string)
	kvStoreMutex = sync.RWMutex{}

	http.HandleFunc("/get", get)
	http.HandleFunc("/set", set)
	http.HandleFunc("/remove", remove)
	http.HandleFunc("/list", list)
	http.ListenAndServe(":3000", nil)

}

// getValue by key
func get(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error", err)
			return
		}
		if len(values.Get("key")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error", "Wrong input key")
			return
		}
		kvStoreMutex.RLock()
		value := keyValueStore[string(values.Get("key"))]
		kvStoreMutex.Unlock()
		fmt.Fprint(w, value)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only GET accepted.")
	}
}

// set value
func set(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error", err)
			return
		}
		if len(values.Get("key")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error", "Wroing input key.")
			return
		}
		if len(values.Get("value")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error", "Wrong input value.")
			return
		}
		kvStoreMutex.Lock()
		keyValueStore[string(values.Get("key"))] = string(values.Get("value"))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted")
	}
}

func remove(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", "Wrong input key.")
			return
		}
		if len(values.Get("key")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", "wrong input key.")
		}

		kvStoreMutex.Lock()
		delete(keyValueStore, values.Get("key"))
		kvStoreMutex.Unlock()
		fmt.Fprint(w, "success")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only DELETE accepted")
	}
}

func list(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		kvStoreMutex.RLock()
		for key, value := range keyValueStore {
			fmt.Fprintln(w, key, ":", value)
		}
		kvStoreMutex.RUnlock()
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only Get accepted.")
	}
}
