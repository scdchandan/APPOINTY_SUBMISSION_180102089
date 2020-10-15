package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"sync"
	"io/ioutil"
	"time"
	"strings"
	"log"
)

type Article struct{
	Id 			string `json:"id"`
	Title 		string `json:"title"`
	SubTitle 	string `json:"subtitle"`
	Content 	string `json:"content"`
	TimeStamp 	string `json:"timestamp"`
}

type articlesHandler struct{
	sync.Mutex
	store map[string]Article
}
func (h *articlesHandler ) ManageArticles (w http.ResponseWriter, r *http.Request){
	
	switch r.Method{
	case "GET":
		h.GetArticles(w,r)
		return
	case "POST":
		h.PostArticle(w,r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method Not Allowed"))
		return
	}
}
func (h *articlesHandler ) GetArticles (w http.ResponseWriter, r *http.Request){
	
	articles := make([]Article, len(h.store))

	h.Lock()
	i := 0
	for _, article := range h.store {
		articles[i] = article
		i++
	}
	h.Unlock()

	jsonBytes, err := json.Marshal(articles)

	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
func (h *articlesHandler ) GetArticlebyId (w http.ResponseWriter, r *http.Request){
	
	parts := strings.Split(r.URL.String(), "/")

	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h.Lock()
	article, ok := h.store[parts[2]]
	h.Unlock()

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}


	jsonBytes, err := json.Marshal(article)

	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
func (h *articlesHandler ) SearchArticlebyQuery (w http.ResponseWriter, r *http.Request){
	
	parts := strings.ToLower(strings.TrimPrefix(r.URL.String(), "/articles/search?q="))
	if len(parts) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	articles := make([]Article, 0)

	h.Lock()
	i := 0
	for _, article := range h.store {

		if strings.Contains(strings.ToLower(article.Title),parts) || strings.Contains(strings.ToLower(article.SubTitle),parts) || strings.Contains(strings.ToLower(article.Content),parts){
			articles = append(articles,article)
			i++
		}
		
	}
	h.Unlock()

	if i == 0 {
		fmt.Println("NOTFOUND")
		w.WriteHeader(http.StatusNotFound)
		return
	}


	jsonBytes, err := json.Marshal(articles)

	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
func (h *articlesHandler ) PostArticle (w http.ResponseWriter, r *http.Request){
	
	bodyBytes, err := ioutil.ReadAll(r.Body)
	log.Print(string(bodyBytes))
	defer r.Body.Close()
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// ct := r.Header.Get("content-type")
	// if ct != "applicatio/json"{
	// 	w.WriteHeader(http.StatusUnsupportedMediaType)
	// 	w.Write([]byte(fmt.Sprintf("need content type 'application/json', got '%s' ", ct)))
	// 	return
	// }

	var article Article
	err = json.Unmarshal(bodyBytes, &article)
	if err != nil{
		
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		// w.Write([]byte(article))
		return
	}
	article.Id = fmt.Sprintf("%d", time.Now().UnixNano())
	h.Lock()
	h.store[article.Id] = article
	defer h.Unlock()
}
func newArticlesHandler() *articlesHandler{
	return &articlesHandler{
		store: map[string]Article{
			"id1": Article{
				Id: 		"id1",
				Title: 		"Greenland",
				SubTitle: 	"Scale of melting ICE in Greenland",
				Content: 	"Greenland's ice sheet has melted to a point of no return, and efforts to slow global warming will not stop it from disintegrating. That's according to a new study by researchers at Ohio State University.",
				TimeStamp: 	"2nd January 2020",
			},
			"id2": Article{
				Id: 		"id2",
				Title: 		"PUNE",
				SubTitle: 	"Heavy rain to continue over Pune today: IMD",
				Content: 	"Pune city recorded 19.8 mm of rain in a three-hour span till 8.30 pm on Wednesday as the sky remained cloudy throughout the day. The rainfall intensity picked up by late evening as the well-marked low pressure system approached over south Maharashtra, causing widespread but moderate-intensity rain.",
				TimeStamp: 	"1st February 2020",
			},
		},
	}
}

func main() {
	articlesHandler := newArticlesHandler()
	http.HandleFunc("/articles", articlesHandler.ManageArticles)
	http.HandleFunc("/articles/", articlesHandler.GetArticlebyId)
	http.HandleFunc("/articles/search", articlesHandler.SearchArticlebyQuery)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}