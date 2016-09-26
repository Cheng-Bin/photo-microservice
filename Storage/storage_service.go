/**
*
* 存储图片微服务，然后使用键值界面和参数表明图片是修改前还是修改后。
* 为了简洁起见，我们把图片保存在一个以图片状态（完成／进行中）命名的文件夹。
*
**/
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

func main() {

	if !registerInKVStore() {
		return
	}

	http.HandleFunc("/sendImg", receiveImage)
	http.HandleFunc("/getImg", serveImage)
	http.ListenAndServe(":3002", nil)
}

// register service to keyvalue
func registerInKVStore() bool {
	if len(os.Args) < 3 {
		fmt.Println("Error: Too few arguments.")
		return false
	}

	storgeAddress := os.Args[1]
	keyValueStoreAddress := os.Args[2]

	response, err := http.Post("http://"+keyValueStoreAddress+"/set?key=storgeAddress&value="+storgeAddress, "", nil)
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
		fmt.Println("Error:", string(data))
		return false
	}

	return true
}

func serveImage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: ", err)
			return
		}
		if len(values.Get("id")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: ", "Wrong input id.")
			return
		}
		if values.Get("state") != "working" && values.Get("state") != "finished" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", "Wrong input state.")
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted")
	}

}

func receiveImage(w http.ResponseWriter, r *http.Request) {

}
