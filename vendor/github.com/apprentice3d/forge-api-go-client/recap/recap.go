package recap

import (
	"errors"
	"log"

	"github.com/apprentice3d/forge-api-go-client/oauth"
)

func NewReCapAPIWithCredentials(ClientID string, ClientSecret string) ReCapAPI {
	recapAPI := ReCapAPI{}
	recapAPI.BasePath = "/photo-to-3d/v1"
	return ReCapAPI{
		oauth.NewTwoLeggedClient(ClientID, ClientSecret),
		"/photo-to-3d/v1",
	}
}

// CreatePhotoScene is used to prepare a scene with a given name and expected output formats
func (api ReCapAPI) CreatePhotoScene(name string, formats []string) (scene PhotoScene, err error) {

	bearer, err := api.Authenticate("data:write")
	if err != nil {
		return
	}
	path := api.Host + api.ReCapPath
	scene, err = CreatePhotoScene(path, name, formats, bearer.AccessToken)

	return
}

func (api ReCapAPI) AddFilesToScene(scene *PhotoScene, files []string) (uploads []FileUploadingReply, err error) {
	bearer, err := api.Authenticate("data:write")
	if err != nil {
		return
	}
	scene.Files = append(scene.Files, files...)
	path := api.Host + api.ReCapPath

	/******** Parallel way ***************/
	//create a channel from which workers will consume
	workChan := make(chan string, len(files))
	for _, filename := range scene.Files {
		workChan <- filename
	}
	close(workChan)

	// since some OS have limits on open file descriptor
	// we have to limit the number of goroutines opening files
	workers := 16
	if len(scene.Files) < workers {
		workers = len(scene.Files)
	}

	successChan := make(chan *FileUploadingReply, len(scene.Files))
	errChan := make(chan error, 1)

	for workerID := 0; workerID < workers; workerID++ {
		go func() {
			for file := range workChan {
				reply, err := AddFileToScene(path, scene.ID, file, bearer.AccessToken)
				if err != nil {
					errChan <- err
					return
				}
				successChan <- &reply
			}
		}()
	}

	for i := 0; i < len(scene.Files); i++ {
		select {
		case result := <-successChan:
			uploads = append(uploads, *result)
			log.Printf("[%d/%d] SUCCESS uploading image: %s\n",
				i+1,
				len(scene.Files),
				result.Files.File.FileName)
		case err = <-errChan:
			return
		}
	}
	/******** END of Parallel way ***************/

	/******** Sequential way ***************/
	//for _, file := range scene.Files {
	//	reply, err := AddFileToScene(path, scene.ID, file, bearer.AccessToken)
	//	if err != nil {
	//		break
	//	}
	//	uploads = append(uploads, reply)
	//}
	/******** END of Sequential way ***************/

	return
}

func (api ReCapAPI) StartSceneProcessing(scene PhotoScene) (sceneID string, err error) {
	bearer, err := api.Authenticate("data:write")
	if err != nil {
		return
	}
	path := api.Host + api.ReCapPath
	sceneID, err = StartSceneProcessing(path, scene.ID, bearer.AccessToken)
	return
}

func (api ReCapAPI) GetSceneProgress(scene PhotoScene) (progress SceneProgressReply, err error) {
	bearer, err := api.Authenticate("data:read")
	if err != nil {
		return
	}
	path := api.Host + api.ReCapPath
	progress, err = GetSceneProgress(path, scene.ID, bearer.AccessToken)
	return
}

func (api ReCapAPI) GetSceneResults(scene PhotoScene, format string) (result SceneResultReply, err error) {
	bearer, err := api.Authenticate("data:read")
	if err != nil {
		return
	}
	path := api.Host + api.ReCapPath
	result, err = GetScene(path, scene.ID, bearer.AccessToken, format)
	return
}

func (api ReCapAPI) CancelSceneProcessing(scene PhotoScene) (sceneID string, err error) {
	err = errors.New("method not implemented")
	return
}

func (api ReCapAPI) DeleteScene(scene PhotoScene) (sceneID string, err error) {
	bearer, err := api.Authenticate("data:write")
	if err != nil {
		return
	}
	path := api.Host + api.ReCapPath
	_, err = DeleteScene(path, scene.ID, bearer.AccessToken)
	sceneID = scene.ID
	return
}
