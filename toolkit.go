package crowdin

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Publicly available high level functions generally combining several API calls

const polldelaysec = 5 // Defines delay between each api call when polling a progress status

// Lookup buildId for current project
func (crowdin *Crowdin) GetBuildId() (buildId int, err error) {

	crowdin.log("GetBuildId()")

	var opt ListProjectBuildsOptions
	rl, err := crowdin.ListProjectBuilds(&opt)
	if err != nil {
		return 0, err
	}
	for _, v := range rl.Data {
		if (v.Data.ProjectId == crowdin.config.projectId) && (v.Data.Status == "finished") {
			buildId = v.Data.Id
		}
	}
	if buildId == 0 {
		return 0, errors.New("Can't find a build for this project or build is in progress.")
	}
	return buildId, nil
}

// Lookup projectId
func (crowdin *Crowdin) GetProjectId(projectName string) (projectId int, err error) {

	crowdin.log("GetProjectId()")

	var opt ListProjectsOptions
	rl, err := crowdin.ListProjects(&opt)
	if err != nil {
		return 0, err
	}

	for _, v := range rl.Data {
		if v.Data.Name == projectName {
			projectId = v.Data.Id
		}
	}
	if projectId == 0 {
		return 0, errors.New("Can't find project.")
	}
	return projectId, nil
}

// BuildAllLg - Build a project for all languages
// Update buildProgress
func (crowdin *Crowdin) BuildAllLg(buildTOinSec int) (buildId int, err error) {
	crowdin.log("BuildAllLg()")

	// Invoke build
	var bo BuildProjectOptions
	// bo.ProjectId = crowdin.config.projectId
	bo.BranchId = 0
	bo.Languages = nil
	rb, err := crowdin.BuildProject(&bo)
	if err != nil {
		return buildId, errors.New("\nBuild Err.")
	}
	buildId = rb.Data.Id
	crowdin.log(fmt.Sprintf("	BuildId=%d", buildId))

	// Poll build status with a timeout
	crowdin.log("	Poll build status crowdin.GetBuildProgress()")
	timer := time.NewTimer(time.Duration(buildTOinSec) * time.Second)
	defer timer.Stop()
	rp := &ResponseGetBuildProgress{}
	for rp.Data.Status = rb.Data.Status; rp.Data.Status != "finished" && rp.Data.Status != "canceled"; { // initial value is read from previous API call
		time.Sleep(polldelaysec * time.Second) // delay between each call
		rp, err = crowdin.GetBuildProgress(&GetBuildProgressOptions{BuildId: buildId})
		if err != nil {
			break
		}
		select {
		case <-timer.C:
			err = errors.New("Build Timeout.")
			return buildId, err
		default:
		}
	}

	if rp.Data.Status != "finished" {
		err = errors.New(fmt.Sprintf("	Build Error:%s", rp.Data.Status))
	}
	return buildId, err
}

// Download a build of the current project
//    outputFileNamePath  required
//    projectId           required if projectName is not provided
//    buildId             optional
// limitation: total number of project directories needs to be 500 max
func (crowdin *Crowdin) DownloadBuild(outputFileNamePath string, buildId int) (err error) {

	// Get URL for downloading
	rd, err := crowdin.DownloadProjectTranslations(&DownloadProjectTranslationsOptions{buildId})
	if err != nil {
		return errors.New("DownloadBuild() - Error getting URL for download.")
	}
	url := rd.Data.Url

	// Actual downloading
	err = crowdin.DownloadFile(url, outputFileNamePath)

	return err
}

// Update a file of the current project
//    outputFileNamePath  required
//    crowdinFileNamePath required
func (crowdin *Crowdin) UpdateFile(localFileNamePath string, crowdinFileNamePath string) (err error) {


	// Lookup fileId in Crowdin
	var dirId int
	crowdinFile := strings.Split(localFileNamePath, "/")
	switch l := len(crowdinFile) {
	case l = 0:
		return errors.New("UpdateFile() - Crowdin file name should not be null.")
	case l = 1: // no directory so dirId is null
	case l > 1: //
		// Lookup end directoryId
		// Get a list of all the project folders
		listDir, err := crowdin.ListDirectories(&ListDirectoriesOptions{Limit: 500})
		if err != nil {
			return errors.New("UpdateFile() - Error listing project directories.")
		}
	}

	if len(listDir.Data) > 0 {
		// Lookup end directoryId
		for i, v := range listDir.Data {
			fmt.Printf("val[%d]= %d, %s %d\n", i, v.Data.Id, v.Data.Name, v.Data.DirectoryId)
		}
	} // else dirId is null

	fmt.Printf( crowdinFile[0])

	return err

}
