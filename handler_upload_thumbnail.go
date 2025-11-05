package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoID_string := r.PathValue("videoID")
	videoID, err := uuid.Parse(r.PathValue("videoID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid video ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing bearer token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT", err)
		return
	}

	fmt.Println("Uploading thumbnail for video:", videoID, "by user:", userID)

	const maxMemory = 10 << 20
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid form data", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Missing thumbnail file", err)
		return
	}
	defer file.Close()

	//_, err = io.ReadAll(file)
	// if err != nil {
	// 	respondWithError(w, http.StatusInternalServerError, "Failed to read thumbnail", err)
	// 	return
	// }

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Video not found", err)
		return
	}

	fmt.Println("----------assetlogic start-------------------")

	//encodedImage := base64.StdEncoding.EncodeToString(tndata)
	fileextension := header.Header.Get("Content-Type")
	slice := strings.Split(fileextension, "/")
	fileextension = slice[1]
	fmt.Println(fileextension)

	thumbnailPath := (filepath.Join(cfg.assetsRoot, videoID_string+"."+fileextension))
	thumbnail, err := os.Create(thumbnailPath)

	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = io.Copy(thumbnail, file)
	if err != nil {
		fmt.Println(err)
		return
	}
	//	dataURL := "data:" + header.Header.Get("Content-Type") + ";base64," + encodedImage
	assetURL := "http://localhost:" + cfg.port + "/" + thumbnailPath
	// n := len(dataURL)
	// fmt.Println(n)
	// fmt.Println(dataURL[n-5:])
	// fmt.Println(dataURL[:30])

	fmt.Println("----------assetlogic end-------------------")

	fmt.Println(assetURL)

	video.ThumbnailURL = &assetURL

	if err := cfg.db.UpdateVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update video thumbnail", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
