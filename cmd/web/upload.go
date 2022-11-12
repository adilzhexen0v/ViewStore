package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/drive/v3"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
)

func (app *application) ServiceAccount(secretFile string) *http.Client {
	b, err := ioutil.ReadFile(secretFile)
	if err != nil {
		log.Fatal("error while reading the credential file", err)
	}
	var s = struct {
		Email      string `json:"client_email"`
		PrivateKey string `json:"private_key"`
	}{}
	json.Unmarshal(b, &s)
	config := &jwt.Config{
		Email:      s.Email,
		PrivateKey: []byte(s.PrivateKey),
		Scopes: []string{
			drive.DriveScope,
		},
		TokenURL: google.JWTTokenURL,
	}
	client := config.Client(context.Background())
	return client
}

func (app *application) createFile(service *drive.Service, name string, mimeType string, content multipart.File, parentId string) (*drive.File, error) {
	f := &drive.File{
		MimeType: mimeType,
		Name:     name,
		Parents:  []string{parentId},
	}
	file, err := service.Files.Create(f).Media(content).Do()

	if err != nil {
		log.Println("Could not create file: " + err.Error())
		return nil, err
	}

	return file, nil
}

func (app *application) uploadFile(filename string, file multipart.File) string {
	client := app.ServiceAccount("./cmd/json/client_secret.json")

	srv, err := drive.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve drive Client %v", err)
	}

	folderId := "1MOltMtzLXaKlj8WMZYBA2S90pHdLn3Cd"

	// Step 4: create the file and upload
	f, err := app.createFile(srv, filename, "image/png", file, folderId)
	if err != nil {
		panic(fmt.Sprintf("Could not create file: %v\n", err))
	}
	return f.Id
}

/*
func main() {
	// Step 1: Open  file
	f, err := os.Open("upload-1372565739.png")

	if err != nil {
		panic(fmt.Sprintf("cannot open file: %v", err))
	}
	fmt.Printf("%T\n", f)
	defer f.Close()

	// Step 2: Get the Google Drive service
	client := ServiceAccount("./cmd/web/client_secret.json")

	srv, err := drive.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve drive Client %v", err)
	}

	folderId := "1MOltMtzLXaKlj8WMZYBA2S90pHdLn3Cd"

	// Step 4: create the file and upload
	file, err := createFile(srv, f.Name(), "image/png", f, folderId)

	if err != nil {
		panic(fmt.Sprintf("Could not create file: %v\n", err))
	}

	fmt.Printf("File '%s' successfully uploaded", file.Name)
	fmt.Printf("\nFile Id: '%s' ", file.Id)

}
*/
