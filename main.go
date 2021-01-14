package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type Post struct {
	Author  string
	Message string
	Posted  time.Time
}

type templateParams struct {
	Notice  string
	Name    string
	Message string
	Posts   []Post
}

func main() {
	router := gin.Default()

	router.LoadHTMLGlob("template/*")

	router.GET("/", indexGetHandle)
	router.POST("/", indexPostHandle)

	http.Handle("/", router)
	appengine.Main()
}

func indexGetHandle(c *gin.Context) {

	ctx := appengine.NewContext(c.Request)
	params := templateParams{}

	q := datastore.NewQuery("Post").Order("-Posted").Limit(20)
	if _, err := q.GetAll(ctx, &params.Posts); err != nil {
		log.Errorf(ctx, "Getting posts: %v", err)
		params.Notice = "Couldn't get latest posts. Refresh?"

		c.HTML(http.StatusInternalServerError, "top/index", gin.H{
			"params": params,
		})
		return
	}

	c.HTML(http.StatusOK, "top/index", gin.H{
		"params": params,
	})
}

func indexPostHandle(c *gin.Context) {

	ctx := appengine.NewContext(c.Request)
	params := templateParams{}

	post := Post{
		Author:  c.PostForm("name"),
		Message: c.PostForm("message"),
		Posted:  time.Now(),
	}

	if post.Author == "" {
		post.Author = "ナナシさん"
	}
	params.Name = post.Author

	if post.Message == "" {
		params.Notice = "No message provided"

		c.HTML(http.StatusBadRequest, "top/index", gin.H{
			"params": params,
		})
		return
	}

	key := datastore.NewIncompleteKey(ctx, "Post", nil)

	if _, err := datastore.Put(ctx, key, &post); err != nil {
		log.Errorf(ctx, "datastore.Put: %v", err)

		params.Notice = "Couldn't add new post. Try again?"
		params.Message = post.Message

		c.HTML(http.StatusInternalServerError, "top/index", gin.H{
			"params": params,
		})
		return
	}

	params.Posts = append([]Post{post}, params.Posts...)

	params.Notice = fmt.Sprintf("Thank you for your submission, %s!", post.Author)

	c.HTML(http.StatusOK, "top/index", gin.H{
		"params": params,
	})
}