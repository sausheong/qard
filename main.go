package main

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"image"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"github.com/emersion/go-vcard"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	qrcode "github.com/yeqown/go-qrcode"
)

var dir string

func init() {
	var err error
	dir, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/makeform", makeForm)
	r.HandleFunc("/make", makeQRCode)
	r.HandleFunc("/sw.js", serviceWorker)
	r.HandleFunc("/manifest.json", manifest)

	r.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir+"/static"))))

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:9000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Starting Qard server at", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

// front page
func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles(dir + "/static/index.html")
	t.Execute(w, nil)
}

func makeForm(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles(dir + "/static/makeform.html")
	t.Execute(w, nil)
}

func makeQRCode(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles(dir + "/static/show.html")
	r.ParseMultipartForm(8192)

	card := make(vcard.Card)
	var firstName, lastName string
	qroptions := []qrcode.ImageOption{}
	for k, v := range r.PostForm {
		if len(v) > 0 {
			switch k {
			case "first_name":
				firstName = v[0]
			case "last_name":
				lastName = v[0]
			case "formatted_name":
				card.SetValue(vcard.FieldFormattedName, v[0])
			case "mobile":
				mobile := vcard.Field{
					Value:  v[0],
					Params: vcard.Params{"TYPE": {"CELL", "VOICE"}},
				}
				card.Add("TEL", &mobile)
			case "office":
				office := vcard.Field{
					Value:  v[0],
					Params: vcard.Params{"TYPE": {"WORK", "VOICE"}},
				}
				card.Add("TEL", &office)
			case "email":
				email := vcard.Field{
					Value:  v[0],
					Params: vcard.Params{"TYPE": {"WORK"}},
				}
				card.Add("EMAIL", &email)
			case "org":
				card.SetValue(vcard.FieldOrganization, v[0])
			case "designation":
				card.SetValue(vcard.FieldTitle, v[0])
			case "url":
				card.SetValue(vcard.FieldURL, v[0])
			case "color":
				qroptions = append(qroptions, qrcode.WithFgColorRGBHex(v[0]))
			}
		}
	}

	file, _, err := r.FormFile("logo")
	var hasLogo bool
	var pic image.Image
	if err == nil {
		b1 := make([]byte, 512)
		_, err = file.Read(b1)
		if err == nil {
			filetype := http.DetectContentType(b1)
			if filetype == "image/jpeg" || filetype == "image/jpg" || filetype == "image/png" {
				file.Seek(0, io.SeekStart)
				pic, _, err = image.Decode(file)
				if err == nil {
					hasLogo = true
				} else {
					log.Println("Cannot decode file - ", err)
				}
			} else {
				log.Println("Not an image file - ", filetype)
			}
		} else {
			log.Println("Cannot read file - ", err)
		}
	}

	if hasLogo {
		logo := resize.Resize(244, 244, pic, resize.Lanczos3)
		qroptions = append(qroptions, qrcode.WithLogoImage(logo))
	} else {
		qroptions = append(qroptions, qrcode.WithLogoImageFilePNG("static/img/transparent.png"))
	}

	name := vcard.Name{
		FamilyName: lastName,
		GivenName:  firstName,
	}
	card.AddName(&name)
	card.SetValue("VERSION", "3.0")

	var cardbuff bytes.Buffer
	enc := vcard.NewEncoder(&cardbuff)
	err = enc.Encode(card)
	if err != nil {
		log.Println("cannot encode card - ", err)
	}

	qrc, err := qrcode.New(cardbuff.String(), qroptions...)
	if err != nil {
		log.Printf("could not generate QRCode: %v", err)
	}

	var qrbuff bytes.Buffer
	if err := qrc.SaveTo(&qrbuff); err != nil {
		log.Printf("could not save image: %v", err)
	}
	qrbase64 := base64.StdEncoding.EncodeToString(qrbuff.Bytes())
	t.Execute(w, qrbase64)
}

func serviceWorker(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("sw.js")
	if err != nil {
		http.Error(w, "Couldn't read file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Write(data)
}

func manifest(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("manifest.json")
	if err != nil {
		http.Error(w, "Couldn't read file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}
