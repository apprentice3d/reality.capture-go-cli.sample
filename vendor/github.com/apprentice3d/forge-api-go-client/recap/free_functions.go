package recap

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//const (
//	basePath = "https://developer.api.autodesk.com/photo-to-3d/v1"
//)

func CreatePhotoScene(path string, name string, formats []string, token string) (scene PhotoScene, err error) {

	task := http.Client{}

	body := url.Values{}
	body.Add("scenename", name)
	body.Add("format", strings.Join(formats, " "))

	req, err := http.NewRequest("POST",
		path+"/photoscene",
		bytes.NewBufferString(body.Encode()),
	)

	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	response, err := task.Do(req)
	if err != nil {
		return
	}

	content, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	if response.StatusCode != 200 {
		err = errors.New("[" + strconv.Itoa(response.StatusCode) + "] " + string(content))
		return
	}

	if err = checkMessageForErrors(content); err != nil {
		return
	}

	sceneCreationReply := SceneCreationReply{}
	err = json.Unmarshal(content, &sceneCreationReply)
	scene.ID = sceneCreationReply.PhotoScene.ID

	return

}

func AddFileToScene(path string, photoSceneId string, filename string, token string) (result FileUploadingReply, err error) {

	if result, err = readFileAndUpload(path, photoSceneId, filename, token); err != nil {
		// Warning: Assuming that if failed to read from localfile, then it is a link
		// TODO: fix this bug for case when local file has wrong path or filename
		result, err = readLinkAndUpload(path, photoSceneId, filename, token)
	}

	return
}

func StartSceneProcessing(path string, photoSceneId string, token string) (sceneID string, err error) {
	task := http.Client{}

	req, err := http.NewRequest("POST",
		path+"/photoscene/"+photoSceneId,
		nil,
	)

	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	response, err := task.Do(req)
	if err != nil {
		return
	}

	content, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	if response.StatusCode != 200 {
		err = errors.New("[" + strconv.Itoa(response.StatusCode) + "] " + string(content))
		return
	}

	if err = checkMessageForErrors(content); err != nil {
		return
	}

	sceneStartProcessingReply := SceneStartProcessingReply{}
	err = json.Unmarshal(content, &sceneStartProcessingReply)
	sceneID = sceneStartProcessingReply.PhotoScene.ID

	return
}

func GetSceneProgress(path string, photoSceneId string, token string) (progress SceneProgressReply, err error) {
	task := http.Client{}

	req, err := http.NewRequest("GET",
		path+"/photoscene/"+photoSceneId+"/progress",
		nil,
	)

	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	response, err := task.Do(req)
	if err != nil {
		return
	}

	content, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	if response.StatusCode != 200 {
		err = errors.New("[" + strconv.Itoa(response.StatusCode) + "] " + string(content))
		return
	}

	if err = checkMessageForErrors(content); err != nil {
		return
	}
	err = json.Unmarshal(content, &progress)

	return
}

func GetScene(path string, photoSceneId string, token string, format string) (result SceneResultReply, err error) {
	task := http.Client{}

	body := strings.NewReader("format=" + format)

	req, err := http.NewRequest("GET",
		path+"/photoscene/"+photoSceneId,
		body,
	)

	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	response, err := task.Do(req)
	if err != nil {
		return
	}

	content, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	if response.StatusCode != 200 {
		err = errors.New("[" + strconv.Itoa(response.StatusCode) + "] " + string(content))
		return
	}

	if err = checkMessageForErrors(content); err != nil {
		return
	}

	err = json.Unmarshal(content, &result)

	return
}

func CancelSceneProcessing(path string, photoSceneId string, token string) (scene PhotoScene, err error) {
	err = errors.New("method not implemented")
	return
}

func DeleteScene(path string, photoSceneId string, token string) (result SceneDeletionReply, err error) {
	task := http.Client{}

	req, err := http.NewRequest("DELETE",
		path+"/photoscene/"+photoSceneId,
		nil,
	)

	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	response, err := task.Do(req)
	if err != nil {
		return
	}

	content, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	if response.StatusCode != 200 {
		err = errors.New("[" + strconv.Itoa(response.StatusCode) + "] " + string(content))
		return
	}

	if err = checkMessageForErrors(content); err != nil {
		return
	}

	err = json.Unmarshal(content, &result)

	return
}

/******************* AUX FUNCTIONS *******************************/

func readFileAndUpload(path string, photoSceneId string, filename string, token string) (result FileUploadingReply, err error) {

	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	formFile, err := writer.CreateFormFile("file[0]", filepath.Base(filename))
	if err != nil {
		log.Println(err.Error())
		return
	}

	if _, err = io.Copy(formFile, file); err != nil {
		log.Println(err.Error())
		return
	}

	writer.WriteField("photosceneid", photoSceneId)
	writer.WriteField("type", "image")
	writer.Close()

	task := http.Client{}

	req, err := http.NewRequest("POST",
		path+"/file",
		body)

	if err != nil {
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := task.Do(req)

	if err != nil {
		return
	}
	defer response.Body.Close()
	content, _ := ioutil.ReadAll(response.Body)

	if response.StatusCode != 200 {
		err = errors.New(response.Request.URL.String() + " => [" + strconv.Itoa(response.StatusCode) + "] " + string(content))
		log.Println(err.Error())
		return
	}

	if err = checkMessageForErrors(content); err != nil {
		return
	}

	err = json.Unmarshal(content, &result)

	return
}

func readLinkAndUpload(path string, photoSceneId string, filename string, token string) (result FileUploadingReply, err error) {
	task := http.Client{}

	body := strings.NewReader(`photosceneid=` + photoSceneId + `&type=image&file[0]=` + filename)

	req, err := http.NewRequest("POST",
		path+"/file",
		body,
	)

	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+token)
	response, err := task.Do(req)
	if err != nil {
		return
	}

	content, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	if response.StatusCode != 200 {
		err = errors.New("[" + strconv.Itoa(response.StatusCode) + "] " + string(content))
		return
	}

	if err = checkMessageForErrors(content); err != nil {
		return
	}

	err = json.Unmarshal(content, &result)

	return

}

// Check if the body is not containing an error message
func checkMessageForErrors(content []byte) (err error) {

	error_checker := ErrorMessage{}
	err = json.Unmarshal(content, &error_checker)

	if error_checker.Error != nil {
		// Got a message containing an error
		err = errors.New("Error " +
			error_checker.Error.Code +
			": " +
			error_checker.Error.Message)
	}
	return
}
