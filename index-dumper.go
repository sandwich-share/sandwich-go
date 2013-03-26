package main

import "io/ioutil"

func main() {
}

func dump(ip string, data string) {
    err := ioutil.WriteFile(ip, []byte(data), 0660 )
    if err != nil {
        panic(err)
    }
}
