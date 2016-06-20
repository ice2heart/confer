package main

//go get gopkg.in/mattes/go-expand-tilde.v1
//go get github.com/google/go-github/github
//go get golang.org/x/oauth2

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
	"gopkg.in/mattes/go-expand-tilde.v1"
)

type jsonobject struct {
	Config map[string]Config
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
func getFileBody(path string) []byte {
	expand, _ := tilde.Expand(path)
	file, err := ioutil.ReadFile(expand)
	if err != nil {
		log.Fatalf("getFileBody fatal: %v", err)
		os.Exit(1)
	}
	return file

}
func getMd5(path string) string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("%s\n", err.Error())
		os.Exit(1)
	}
	md5 := md5.New()
	io.Copy(md5, f)
	sum := md5.Sum(nil)
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

func getClient(token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)
	return client
}

func getGist(client *github.Client, id string) *github.Gist {

	gist, res, err := client.Gists.Get(id)
	if err != nil {
		log.Fatalf("Get gist error %v", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	return gist
}

func addFile(gist *github.Gist, name github.GistFilename, data []byte) {
	newFile := new(github.GistFile)
	newData := string(data)
	newFile.Content = &newData
	sizeData := len(newData)
	newFile.Size = &sizeData
	gist.Files[name] = *newFile
}

func updateGist(client *github.Client, id string, gist *github.Gist) {
	_, _, err := client.Gists.Edit(id, gist)
	if err != nil {
		log.Fatalf("Update gist error %v", err)
		os.Exit(1)
	}
}

var filesData string
var githubKey string
var gistKey string
var metadataName github.GistFilename

func init() {
	flag.StringVar(&filesData, "files", "config.json", "Files to watch")
	flag.StringVar(&githubKey, "key", "", "Key to access gist")
	flag.StringVar(&gistKey, "gist", "", "Id gist storage")
	metadataName = ".metadata"
}

func main() {
	flag.Parse()
	if (gistKey == "") || (githubKey == "") {
		log.Fatal("Need to set gist id and github key")
		os.Exit(1)
	}
	log.Println("Ready for working")
	file := getFileBody(filesData)
	var localMeta jsonobject
	if e := json.Unmarshal(file, &localMeta); e != nil {
		log.Print(e)
		os.Exit(1)
	}
	log.Printf("Data %v", localMeta)

	for key, item := range localMeta.Config {
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
		item.Md5 = getMd5(path)
		item.LastModified = getTimeModified(path)
		localMeta.Config[key] = item
	}
	client := getClient(githubKey)
	gist := getGist(client, gistKey)
	gistMeta := gist.Files[metadataName]

	if gistMeta.Size == nil {
		data, err := json.Marshal(localMeta)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		addFile(gist, metadataName, data)
		for key, value := range localMeta.Config {
			addFile(gist, github.GistFilename(key), getFileBody(value.Path))
		}

	} else {
		var remoteMeta jsonobject
		if e := json.Unmarshal([]byte(*gistMeta.Content), &remoteMeta); e != nil {
			log.Fatal(e)
			os.Exit(1)
		}

		for localKey, v := range localMeta.Config {
			if val, ok := remoteMeta.Config[localKey]; ok {
				if val.LastModified < v.LastModified {
					remoteMeta.Config[localKey] = v
					addFile(gist, github.GistFilename(localKey), getFileBody(v.Path))
				}
			} else {
				remoteMeta.Config[localKey] = v
				addFile(gist, github.GistFilename(localKey), getFileBody(v.Path))
			}
		}
		data, err := json.Marshal(localMeta)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		addFile(gist, metadataName, data)
	}
	log.Print(gist.Files)
	updateGist(client, gistKey, gist)

}
