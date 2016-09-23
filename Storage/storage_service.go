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
	"os"
)

func main() {

	if !registerInKVStore() {
		return
	}

	http.HandleFunc("/sendImg", serImg)
	http.HandleFunc("/getImg", sevImg)
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

func serImg(w http.ResponseWriter, r *http.Request) {

}

func sevImg(w http.ResponseWriter, r *http.Request) {

}
