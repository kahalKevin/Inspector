package model

import (
    "time"

    "github.com/gorilla/websocket"    
)

type User struct {
    Id          string       
    Connection  *websocket.Conn
    TimeIn      time.Time    
}

type Ids []string

func NewUser(id string, connection *websocket.Conn, timein time.Time) User {
  // Create, and return the worker.
  user := User{
    Id:            id,
    Connection:    connection,
    TimeIn:        timein}
  
  return user
}