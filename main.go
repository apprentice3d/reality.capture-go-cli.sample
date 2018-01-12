package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/apprentice3d/forge-api-go-client/recap"
)

func main() {

	dir := "."

	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	images, err := getListOfJPGFilesFromPath(dir)
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Printf("Found %d jpg images.\n", len(images))

	clientID, clientSecret, err := getCredentials()

	if err != nil {
		log.Fatalln(err.Error())
	}

	recapAPI := recap.NewReCapAPIWithCredentials(clientID, clientSecret)

	fmt.Println("Creating a scene ...")
	scene, err := recapAPI.CreatePhotoScene("example", []string{"obj"})
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Printf("Scene created: id = %s\n", scene.ID)

	fmt.Println("Uploading sample images ...standby...")
	uploadResults, err := recapAPI.AddFilesToScene(&scene, images)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Printf("%d files where uploaded sucessfully\n", len(uploadResults))

	fmt.Println("Starting scene processing ...")
	if _, err = recapAPI.StartSceneProcessing(scene); err != nil {
		log.Println(err.Error())
	}

	fmt.Println("Checking scene status ...")
	var progressResult recap.SceneProgressReply
	var ratio float64
	for {
		if progressResult, err = recapAPI.GetSceneProgress(scene); err != nil {
			log.Printf("Failed to get the PhotoScene progress: %s\n", err.Error())
			return
		}

		ratio, _ = strconv.ParseFloat(progressResult.PhotoScene.Progress, 64)
		if err != nil {
			log.Printf("Failed to parse progress results: %s\n", err.Error())
			return
		}

		if ratio == float64(100.0) {
			break
		}
		fmt.Printf("\rScene progress = %.2f%%", ratio)
		time.Sleep(5 * time.Second)
	}

	fmt.Println("\nFinished processing the scene, now getting the results in obj format...")
	result, err := recapAPI.GetSceneResults(scene, "obj")
	if err != nil {
		log.Println(err.Error())
	}

	fmt.Printf("Results are available at following link => %s\n", result.PhotoScene.SceneLink)
	if err := downloadLink(result.PhotoScene.SceneLink); err != nil {
		log.Println("WARNING: Could not download the provided link")
	} else {
		workDir, _ := os.Getwd()
		fmt.Printf("File downloaded to %s as 'result.zip'\n", workDir)
	}

	fmt.Println("Deleting the scene ...")
	_, err = recapAPI.DeleteScene(scene)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("Scene deleted successfully!")
}

func downloadLink(link string) (err error) {
	resp, err := http.Get(link)

	if err != nil {
		return
	}
	defer resp.Body.Close()
	result, err := os.Create("result.zip")
	defer result.Close()
	if err != nil {
		return
	}

	_, err = io.Copy(result, resp.Body)

	return
}

func getListOfJPGFilesFromPath(dir string) (images []string, err error) {
	files, err := ioutil.ReadDir(dir)

	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() {
			if strings.Compare(strings.ToLower(filepath.Ext(file.Name())), ".jpg") == 0 {
				images = append(images, filepath.Join(dir, file.Name()))
			}
		}
	}

	if len(images) == 0 {
		err = errors.New("no valid images found for upload")
	}

	return
}

func getCredentials() (clientID string, clientSecret string, err error) {
	clientID = os.Getenv("FORGE_CLIENT_ID")
	clientSecret = os.Getenv("FORGE_CLIENT_SECRET")

	if len(clientID) == 0 || len(clientSecret) == 0 {
		err = errors.New("\nFORGE_CLIENT_ID and FORGE_CLIENT_SECRET env vars could not be found.\n" +
			"We encourage using Forge secrets by specifying them as env variables.\nExiting ...")
	}

	return
}
