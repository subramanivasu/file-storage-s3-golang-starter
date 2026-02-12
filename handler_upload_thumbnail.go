package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here

	const maxMemmory = 10 << 20

	err = r.ParseMultipartForm(maxMemmory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse multipart form", err)
		return
	}
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	fileExtension:= header.Header.Get("Content-Type")
	fileExtension=fileExtension[6:]
	filePath:=filepath.Join(cfg.assetsRoot,videoIDString)
	filePath=fmt.Sprintf(filePath+".%s",fileExtension)
	
	osFile,err:=os.Create(filePath)
	if err!=nil {
		respondWithError(w, http.StatusBadRequest, "Unable to create new file", err)
		return
	}
	_,err=io.Copy(osFile,file)
	if err!=nil {
		respondWithError(w, http.StatusBadRequest, "Unable to copy file contents", err)
		return
	}
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to Get Video", err)
		return
	}
	if userID != video.UserID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
	}

	thumbnailUrl:=fmt.Sprintf("http://localhost:8091/assets/%s.%s",videoIDString,fileExtension)

	video.ThumbnailURL = &thumbnailUrl
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
