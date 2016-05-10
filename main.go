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
	/*func (b *GistBackend) Push(id string, content []byte) error {
	gist, _, err := b.Client.Gists.Get(id)
	if err != nil {
		return err
	}

	gistFile := gist.Files[gistFilename]
	contentString := string(content)
	gistFile.Content = &contentString
	gist.Files[gistFilename] = gistFile
	log.Debug("Pushing content to Gist backend:", *gistFile.Content)

	_, _, err = b.Client.Gists.Edit(id, gist)
	return err
	}	*/
	log.Print(edit, err, res)
	return gist
}

var filesData string
var githubKey string
var gistKey string

func init() {
	flag.StringVar(&filesData, "files", "config.json", "Files to watch")
	flag.StringVar(&githubKey, "key", "", "Key to access gist")
	flag.StringVar(&gistKey, "gist", "", "Id gist storage")
}

func main() {
	flag.Parse()
	log.Println("Ready for working")
	file, err := ioutil.ReadFile(filesData)
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
		_, e := os.Stat(path)
		if e != nil {
			log.Print("File not found ", item.Path)
			continue
		}
		jsontype.Config[num].Md5 = getMd5(path)
		jsontype.Config[num].LastModified = getTimeModified(path)
	}
	defer printObj(jsontype)
	if (gistKey == "") || (githubKey == "") {
		log.Fatal("Need to set gist id and github key")
		os.Exit(1)
	}
	gist := getGist(githubKey, gistKey)
	log.Print(gist.Files)

}
