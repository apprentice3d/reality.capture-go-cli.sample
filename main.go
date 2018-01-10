package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/apprentice3d/forge-api-go-client/recap"
)

func main() {
	clientID := os.Getenv("FORGE_CLIENT_ID")
	clientSecret := os.Getenv("FORGE_CLIENT_SECRET")

	recapAPI := recap.NewReCapAPIWithCredentials(clientID, clientSecret)

	log.Print("Creating a scene ...")
	scene, err := recapAPI.CreatePhotoScene("example", []string{"obj"})
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Printf("Created a scene with id = %s\n", scene.ID)

	fileSamples := []string{
		"https://s3.amazonaws.com/adsk-recap-public/forge/lion/DSC_1158.JPG",
		"https://s3.amazonaws.com/adsk-recap-public/forge/lion/DSC_1159.JPG",
		"https://s3.amazonaws.com/adsk-recap-public/forge/lion/DSC_1160.JPG",
		"https://s3.amazonaws.com/adsk-recap-public/forge/lion/DSC_1162.JPG",
		"https://s3.amazonaws.com/adsk-recap-public/forge/lion/DSC_1163.JPG",
		"https://s3.amazonaws.com/adsk-recap-public/forge/lion/DSC_1164.JPG",
		"https://s3.amazonaws.com/adsk-recap-public/forge/lion/DSC_1165.JPG",
	}

	log.Print("Uploading sample images ...")
	uploadResults, err := recapAPI.AddFilesToScene(&scene, fileSamples)
	if err != nil {
		log.Fatal(err.Error())
	}
	for idx, result := range uploadResults {
		log.Printf("[%d] Successfully uploaded: %s\n", idx, result.Files.File.FileName)
	}

	log.Print("Starting scene processing ...")
	if _, err = recapAPI.StartSceneProcessing(scene); err != nil {
		log.Println(err.Error())
	}

	log.Print("Checking scene status ...")
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
		log.Printf("\rScene progress = %.2f%% ... waiting 5 seconds", ratio)
		time.Sleep(5 * time.Second)
	}

	log.Print("\nFinished processing the scene, now getting the results ...")
	result, err := recapAPI.GetSceneResults(scene, "obj")
	if err != nil {
		log.Println(err.Error())
	}
	log.Printf("Received the following link %s\n", result.PhotoScene.SceneLink)

	log.Print("Deleting the scene ...")

	_, err = recapAPI.DeleteScene(scene)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Print("Scene deleted successfully!")
}
