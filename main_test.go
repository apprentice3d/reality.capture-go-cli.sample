package main

import (
	"os"
	"testing"
)

func TestJPGFinding(t *testing.T) {
	dir := "./sample_images"
	t.Log(os.Args)
	if len(os.Args) > 1 {
		dir = os.Args[4]
	}

	images, err := getListOfJPGFilesFromPath(dir)
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, image := range images {
		file, err := os.Open(image)
		defer file.Close()
		if err != nil {
			t.Fatalf("Could not open file: %s\n", image)
		} else {
			t.Logf("Image %s opened successfully\n", image)
		}
	}

	if len(images) == 0 {
		t.Fatal("Could not find images")
	}
}
