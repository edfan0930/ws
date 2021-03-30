package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/edfan0930/ws/client"
	wsManager "github.com/edfan0930/ws/manager"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var Manager *wsManager.Manager
var done chan struct{}

type (
	Reciver struct {
		Num  int    `json:"num"`
		Name string `json:"name"`
	}
)

func init() {
	Manager = wsManager.NewManager()

	done = make(chan struct{})
	go Manager.Start(done)

}

func hello(c echo.Context) error {

	id := c.Param("id")

	wsClient, err := client.NewClient(id, c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	Manager.PingErrHandle = func(err error) {
		fmt.Println("is ping err", err)
	}

	Manager.Register(id, wsClient)
	message := wsManager.NewMessage("server", id, []byte("hello client"))
	Manager.Reciver(message)

	for {
		recive := &Reciver{}
		_, err := wsClient.ReadMessage(recive)
		if err != nil {
			fmt.Println("error", err)
		}
		break
	}

	done <- struct{}{}

	return nil
}

//toClient 廣播
func toClient(c echo.Context) error {
	timmer := time.NewTimer(5 * time.Second)
	<-timmer.C

	message := wsManager.NewMessage("server", "all", []byte(`{"Message":"server broadcast"}`))

	fmt.Println(Manager.Reciver(message))
	return c.JSON(http.StatusOK, "done")
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//	e.Static("/", "../public")
	e.GET("/ws/:id", hello)

	e.GET("/toclient", toClient)
	e.Logger.Fatal(e.Start(":1323"))
}
