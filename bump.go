package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gernest/front"
	"github.com/gomarkdown/markdown"
	. "github.com/logrusorgru/aurora"
	"gopkg.in/yaml.v2"
)

const (
	header           = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
	rssSchemaVersion = 2.0
	schemaURL        = "http://www.andymatuschak.org/xml-namespaces/sparkle"
	schemaDc         = "http://purl.org/dc/elements/1.1/"
	releaseDirectory = "./release"
)

type rss struct {
	Channel channel `xml:"channel"`
	Version float32 `xml:"version,attr"`
	Schema  string  `xml:"xmlns:sparkle,attr"`
	Dc      string  `xml:"xmlns:dc,attr"`
}

type channel struct {
	Items []item `xml:"item"`
}

type item struct {
	Enclosure            enclosure    `xml:"enclosure"`
	MinimumSystemVersion *string      `xml:"sparkle minimumSystemVersion"`
	PublishDate          string       `xml:"pubDate"`
	Description          *description `xml:"description"`
}

type enclosure struct {
	URL       string  `xml:"url,attr"`
	Version   string  `xml:"sparkle version,attr"`
	Length    int64   `xml:"length,attr"`
	Type      string  `xml:"type,attr"`
	Signature *string `xml:"sparkle edSignature,attr"`
}

type description struct {
	XMLName xml.Name `xml:"description"`
	Text    string   `xml:",cdata"`
}

type midasConfig struct {
	AppName      string `yaml:"app_name"`
	AppFilename  string `yaml:"app_filename"`
	S3BucketName string `yaml:"s3_bucket_name"`
}

func main() {
	fmt.Println("✨ Running Bump...")

	// Read in frontmatter and markdown from `releaseFile`
	dat, _ := ioutil.ReadFile(findFileWithExtension(".md"))
	m := front.NewMatter()
	m.Handle("---", front.YAMLHandler)
	f, body, err := m.Parse(strings.NewReader(string(dat)))
	htmlDescription := markdown.ToHTML([]byte(body), nil, nil)

	if err != nil {
		fmt.Println("A problem occurred while parsing the release file")
		panic(err)
	}

	zip, err := os.Stat(findFileWithExtension(".zip"))

	if err != nil {
		fmt.Println("A problem occurred while reading your app archive")
		panic(err)
	}

	var signature *string
	sparklePrivateKey := os.Getenv("SPARKLE_PRIVATE_KEY")

	if sparklePrivateKey != "" {
		fmt.Println("Found Sparkle Priate Key at SPARKLE_PRIVATE_KEY")
		newSig := signFileWithKey(sparklePrivateKey)
		signature = &newSig
		fmt.Println("Signed file with private key")
	}

	// Forms the s3 bucket domain
	var urlToVersionedFile string
	var c midasConfig
	c.getConfiguration()
	urlToVersionedFile = fmt.Sprintf("https://%s.s3.amazonaws.com/", c.S3BucketName)

	// Assigns a minimum system version if there is one
	var minimumSystemVersion *string
	if f["minimumSystemVersion"] != nil {
		v := fmt.Sprintf("%s", f["minimumSystemVersion"])
		minimumSystemVersion = &v
	}

	// Turns periods into hyphens
	var urlSafeFilename string
	urlSafeFilename = strings.Replace(fmt.Sprintf("%s", f["version"]), ".", "-", -1)

	// Build our new release
	newItem := item{
		Enclosure: enclosure{
			URL:       fmt.Sprintf("%s%s/%s", urlToVersionedFile, urlSafeFilename, c.AppFilename),
			Version:   fmt.Sprintf("%s", f["version"]),
			Type:      "application/octet-stream",
			Length:    zip.Size(),
			Signature: signature,
		},
		PublishDate:          time.Now().Format("Mon, 01 Jan 2006 15:04:05 +0000"),
		Description:          &description{Text: string(htmlDescription)},
		MinimumSystemVersion: minimumSystemVersion,
	}

	// Find any existing items in our changelog
	existingItems := getExistingChangelogItems()

	// Loop through them to make sure none have the same version number as this release
	for i := 0; i < len(existingItems); i++ {
		cursor := existingItems[i]

		if cursor.Enclosure.Version == newItem.Enclosure.Version {
			existingItems = RemoveRelease(existingItems, i)
		}
	}

	allItems := []item{}
	allItems = append(allItems, newItem)          // Add our newest item first
	allItems = append(allItems, existingItems...) // ...then add existing items

	// Craft our changelog object
	release := &rss{
		Version: rssSchemaVersion,
		Schema:  schemaURL,
		Dc:      schemaDc,
		Channel: channel{
			Items: allItems,
		},
	}

	writeChangelog(*release)
	writeStateFile(newItem.Enclosure.Version)
}

func getExistingChangelogItems() []item {
	xmlFile, err := os.Open(findFileWithExtension(".xml"))
	if err != nil {
		fmt.Println(err)
	}

	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)
	var rss rss
	xml.Unmarshal(byteValue, &rss)

	var items []item

	items = rss.Channel.Items

	return items
}

func writeChangelog(data rss) {
	// Output the marshalled struct
	file, _ := xml.MarshalIndent(data, "", " ")
	err := ioutil.WriteFile(fmt.Sprintf("%s/changelog.xml", releaseDirectory), []byte(header+string(file)), 0644)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(Green("Changelog successfully written"))
}

func writeStateFile(version string) {
	// Write the state file used by Terraform
	file := fmt.Sprintf("version: \"%s\"", version)
	err := ioutil.WriteFile("./.bump/.state/latestRelease", []byte(string(file)), 0644)

	if err != nil {
		fmt.Println(err)
	}
}

func findFileWithExtension(ext string) string {
	var matchingFilePath string
	files, err := ioutil.ReadDir(releaseDirectory)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		extension := filepath.Ext(file.Name())
		if extension == ext {
			matchingFilePath = fmt.Sprintf("%s/%s", releaseDirectory, file.Name())
		}
	}

	return matchingFilePath
}

func (c *midasConfig) getConfiguration() *midasConfig {
	yamlFile, err := ioutil.ReadFile("./config.yml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}

func RemoveRelease(s []item, index int) []item {
	return append(s[:index], s[index+1:]...)
}