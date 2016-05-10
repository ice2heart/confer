package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

type jsonobject struct {
	Config []Config
}

type Config struct {
	Path         string
	Md5          string
	LastModified int
}

/*func getZero() int {
	return 0
}*/

func main() {
	var filesData = flag.String("files", "config.json", "Files to watch")
	flag.Parse()
	log.Println("Ready for working")
	file, e := ioutil.ReadFile(*filesData)
	if e != nil {
		log.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	var jsontype jsonobject
	if err := json.Unmarshal(file, &jsontype); err != nil {
		log.Print(err)
		os.Exit(1)
	}
	log.Printf("Data %v", jsontype)
	/*os.Exit(0)
	dec := json.NewDecoder(os.Stdin)
	enc := json.NewEncoder(os.Stdout)
	for {
		var v map[string]interface{}
		if err := dec.Decode(&v); err != nil {
			log.Println(err)
			return
		}
		for k := range v {
			if k != "Name" {
				delete(v, k)
			}
		}
		if err := enc.Encode(&v); err != nil {
			log.Println(err)
		}
	}*/
}
