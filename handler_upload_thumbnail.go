package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		fmt.Println(err)
		return
	}

	tndata, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Video owner not found", err)
		return
	}

	tn := &thumbnail{
		data:      tndata,
		mediaType: header.Header.Get("Content-Type"),
	}

	// videoThumbnails[video.ID] = *tn
	fmt.Println("---------------")
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	encoded_image := base64.StdEncoding.EncodeToString(tndata)
	fmt.Println(encoded_image[:8])
	fmt.Print("=====================================")

	dataURL := "data:" + tn.mediaType + ";base64," + encoded_image
	// url := "http://localhost:" + port + "/api/thumbnails/" + videoIDString

	fmt.Println(dataURL[:8])
	video.ThumbnailURL = &dataURL

	if err := cfg.db.UpdateVideo(video); err != nil {
		fmt.Println(err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
