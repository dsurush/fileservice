package app

import (
	"errors"
	"github.com/dsurush/fileservice/pkg/core/file"
	"github.com/dsurush/mux/pkg/mux"
	"github.com/dsurush/rest/pkg/rest"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

var (
	ext = make(map[string]string)

	content = `
To get file  type
	GET http://address/file/8f90bec4-abbd-41bf-a86c-41e9744cc03b
`
)
const (
	defaultMaxMemory = 32 << 20 // 32 MB
	contentTypeHtml  = "text/html"
	contentTypeText  = "text/plain"
	contentTypePng   = "image/png"
	contentTypeJpg   = "image/jpeg"
	contentTypePdf   = "application/pdf"
)

type Server struct {
	router  *mux.ExactMux
	fileSvc *file.Service
}

func NewServer(router *mux.ExactMux, fileSvc *file.Service) *Server {
	ext[".txt"] = contentTypeText
	ext[".pdf"] = contentTypePdf
	ext[".png"] = contentTypePng
	ext[".jpg"] = contentTypeJpg
	ext[".html"] = contentTypeHtml
	return &Server{router: router, fileSvc: fileSvc}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.router.ServeHTTP(writer, request)
}

func (s *Server) Stop() {
	//todo: make stop
}

func (s *Server) Start() {
	s.InitRoutes()
}

func (s *Server) handleIndex() http.HandlerFunc {
	tpl, err := template.ParseFiles("index.gohtml")
	if err != nil {
		panic(err)
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		err := tpl.Execute(writer,
			struct {
				Title   string
				Content string
			}{
				Title:   "File Service",
				Content: content,
			})
		if err != nil {
			log.Printf("error while executing template %s %v", tpl.Name(), err)
		}
	}
}

func (s *Server) handleSaveFiles() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		err := request.ParseMultipartForm(defaultMaxMemory)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		files := request.MultipartForm.File["file"]
		type FileURL struct {
			Name string
		}
		fileURLs := make([]FileURL, 0, len(files))

		for _, file := range files {
			contentType, ok := ext[filepath.Ext(file.Filename)]
			if !ok {
				http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			openFile, err := file.Open()
			if err != nil {
				http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			newFile, err := s.fileSvc.Save(openFile, contentType)
			if err != nil {
				http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			fileURLs = append(fileURLs, FileURL{
				newFile[:len(newFile)-len(filepath.Ext(newFile))],
			})
		}
		err = rest.WriteJSONBody(writer, fileURLs)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleGetFile() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		dir, err := ioutil.ReadDir(s.fileSvc.Filepath)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		path := request.URL.Path
		path = path[6:]


		for _, info := range dir {
			if !info.IsDir() {
				fileName := info.Name()
				fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]
				log.Println(fileName, path)
				if !strings.EqualFold(fileName, path){
					continue
				}

				writer.Header().Set("Content-Type", ext[filepath.Ext(info.Name())])
				body, err := ioutil.ReadFile("files/"+info.Name())
				if err != nil {
					http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				_, err = writer.Write(body)
				if err != nil {
					log.Println(errors.New("error"))
				}
				return
			}
		}
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}
