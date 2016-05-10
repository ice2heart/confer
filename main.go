package main

//go get gopkg.in/mattes/go-expand-tilde.v1
//go get github.com/google/go-github/github
//go get golang.org/x/oauth2

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
	"gopkg.in/mattes/go-expand-tilde.v1"
)

type jsonobject struct {
	Config []Config
}

type Config struct {
	Path         string
	Md5          string
	LastModified int64
}

func printObj(obj jsonobject) {
	b, err := json.Marshal(&obj)
	if err != nil {
		log.Print(err)
		return
	}
	log.Printf("%s", b)
}

func getMd5(path string) string {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return ""
	}
	md5 := md5.New()
	io.Copy(md5, f)
	sum := md5.Sum(nil)
	//fmt.Printf("%x\t%s\n", sum, file)
	f.Close()
	return hex.EncodeToString(sum)
}

func getTimeModified(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		log.Print("Error get last Time modified")
	}
	return info.ModTime().Unix()
}

func getGist(token string, id string) *github.Gist {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)

	gist, res, err := client.Gists.Get(id)
	if err != nil {
		log.Fatalf("Get gist error %v", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	edit, res, err := client.Gists.Edit(id, gist) //сюда уже тыкаем обновленный гист
	log.Print(edit, err, res)
	return gist
}

func main() {
	var filesData = flag.String("files", "config.json", "Files to watch")
	flag.Parse()
	log.Println("Ready for working")
	file, err := ioutil.ReadFile(*filesData)
	if err != nil {
		log.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	var jsontype jsonobject
	if e := json.Unmarshal(file, &jsontype); e != nil {
		log.Print(e)
		os.Exit(1)
	}
	log.Printf("Data %v", jsontype)

	for num, item := range jsontype.Config {
		log.Printf("Items %v", item.Path)
		path, err := tilde.Expand(item.Path)
		if err != nil {
			log.Printf("tilde error %v", err)
			os.Exit(1)
		}
		jsontype.Config[num].Md5 = getMd5(path)
		jsontype.Config[num].LastModified = getTimeModified(path)
	}
	defer printObj(jsontype)

	gist := getGist("557f984e0d1d4dbef0270f141b58054ad90fda65", "85ff262891f338732207")
	log.Print(gist.Files)

}
