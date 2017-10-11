package main

import (
  "fmt"
  "os"
  "log"
  "net/http"

  "github.com/garyburd/redigo/redis"
  "github.com/jhunt/vcaptive"
)

func handler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Hi there, sending this message from main.go handler!")
}

func main() {

  fs := http.FileServer(http.Dir("htdocs"))
  http.Handle("/", fs)

  services, err := vcaptive.ParseServices(os.Getenv("VCAP_SERVICES"))

  if err != nil {
    fmt.Fprintf(os.Stderr, "VCAP_SERVICES: %s\n", err)
    os.Exit(1)
  }

  instance, found := services.Tagged("redis")
  if !found {
    fmt.Fprintf(os.Stderr, "VCAP_SERVICES: no 'redis' service found\n")
    os.Exit(2)
  }

 host, ok := instance.GetString("host")
  if !ok {
    fmt.Fprintf(os.Stderr, "VCAP_SERVICES: '%s' service has no 'host' credential\n", instance.Label)
    os.Exit(3)
  }

  port, ok := instance.GetUint("port")
  if !ok {
    fmt.Fprintf(os.Stderr, "VCAP_SERVICES: '%s' service has no 'port' credential\n", instance.Label)
    os.Exit(3)
  }

  password, ok := instance.Get("password")
  if !ok {
    fmt.Fprintf(os.Stderr, "VCAP_SERVICES: '%s' service has no 'password' credential\n", instance.Label)
    os.Exit(3)
  }

  fmt.Printf("Host is %s.  Port is %d.  Password is %s.\n", host, port, password)
  // ...

  conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", host, port), redis.DialPassword(password.(string)))
  if err != nil {
      log.Fatal(err)
  }

  // Importantly, use defer to ensure the connection is always properly
  // closed before exiting the main() function.
  defer conn.Close()

  // Send our command across the connection.
  _, err = conn.Do("APPEND", "key-example", "value-example")

  // Check the Err field.
  if err != nil {
      log.Fatal(err)
  }

  fmt.Println("key value added!")

// http.HandleFunc("/", handler)
// http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil)

  log.Println("Listening...")
  // http.ListenAndServe(":3000", nil)
  http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil)
}
